package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/db/mysqlutil"
	"github.com/1nterdigital/aka-im-tools/db/redisutil"
	"github.com/1nterdigital/aka-im-tools/discovery/etcd"
	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/system/program"
	"github.com/1nterdigital/aka-im-tools/utils/datautil"
	"github.com/1nterdigital/aka-im-tools/utils/network"
	"github.com/1nterdigital/aka-im-tools/utils/runtimeenv"
	walletmw "github.com/1nterdigital/aka-im-wallet/internal/api/mw"
	"github.com/1nterdigital/aka-im-wallet/internal/api/util"
	"github.com/1nterdigital/aka-im-wallet/internal/repository"
	"github.com/1nterdigital/aka-im-wallet/internal/service"
	"github.com/1nterdigital/aka-im-wallet/internal/usecase"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/config"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/database"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/imapi"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/kdisc"
	disetcd "github.com/1nterdigital/aka-im-wallet/pkg/common/kdisc/etcd"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/tokenverify"
)

type Config struct {
	ApiConfig      config.API
	Discovery      config.Discovery
	Share          config.Share
	Admin          config.Admin
	RedisConfig    config.Redis
	PostgresConfig config.Postgres
	MysqlConfig    config.Mysql
	KafkaConfig    config.Kafka
	TracerConfig   config.Tracer
	RuntimeEnv     string
}

type walletService struct {
	Database database.WalletDatabaseInterface
	Token    *tokenverify.Token
}

func initDatabase(
	ctx context.Context, cfg *Config,
) (conn *gorm.DB, mysqlClient *mysqlutil.Client, err error) {
	pgDB, err := mysqlutil.NewMysqlDB(ctx, cfg.MysqlConfig.Build())
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: pgDB.DB, // wrap existing *sql.DB
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	if err = db.InitiateTable(gormDB); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return gormDB, pgDB, nil
}

// initService wires up repository, usecase, and wallet service
func initService(
	cfg *Config, conn *gorm.DB, mysqlDB *mysqlutil.Client, rdb redis.UniversalClient,
) (*walletService, *usecase.UseCase, error) {
	repo := repository.NewRepository(conn)

	uc, err := usecase.New(&usecase.Config{KafkaConfig: cfg.KafkaConfig}, repo, conn)
	if err != nil {
		return nil, nil, err
	}

	srv := &walletService{
		Token: &tokenverify.Token{
			Expires: time.Duration(cfg.Admin.TokenPolicy.Expire) * 24 * time.Hour,
			Secret:  cfg.Admin.Secret,
		},
	}

	srv.Database, err = database.NewWalletDatabase(mysqlDB, rdb, srv.Token)
	if err != nil {
		return nil, nil, err
	}

	return srv, uc, nil
}

// setupServer configures Gin + HTTP server
func setupServer(cfg *Config, walletApi *service.Api, mwApi *walletmw.MW, apiPort int) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	engine := SetRouter(cfg.TracerConfig.AppName.Api, walletApi, mwApi)

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", apiPort),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func Start(ctx context.Context, index int, cfg *Config) error {
	log.CInfo(ctx, "Starting WALLET-API server instance")
	cfg.RuntimeEnv = runtimeenv.PrintRuntimeEnvironment()

	if len(cfg.Share.WalletAdmin) == 0 {
		return errs.New("share wallet admin not configured")
	}

	rdb, err := redisutil.NewRedisClient(ctx, cfg.RedisConfig.Build())
	if err != nil {
		return err
	}

	conn, pgDB, err := initDatabase(ctx, cfg)
	if err != nil {
		return err
	}

	srv, uc, err := initService(cfg, conn, pgDB, rdb)
	if err != nil {
		return err
	}

	client, err := kdisc.NewDiscoveryRegister(&cfg.Discovery, cfg.RuntimeEnv, nil)
	if err != nil {
		return err
	}

	im := imapi.New(cfg.Share.AkaIM.ApiURL, cfg.Share.AkaIM.Secret, cfg.Share.AkaIM.AdminUserID)
	base := util.Api{
		ImUserID:          cfg.Share.AkaIM.AdminUserID,
		ProxyHeader:       cfg.Share.ProxyHeader,
		WalletAdminUserID: cfg.Share.WalletAdmin[0],
	}

	walletApi := service.New(im, &base, uc)
	mwApi := walletmw.New(srv.Token, srv.Database)

	apiPort, err := datautil.GetElemByIndex(cfg.ApiConfig.Api.Ports, index)
	if err != nil {
		return err
	}
	address := net.JoinHostPort(network.GetListenIP(cfg.ApiConfig.Api.ListenIP), strconv.Itoa(apiPort))
	server := setupServer(cfg, walletApi, mwApi, apiPort)

	log.CInfo(ctx, "API server is initializing", "address", address, "apiPort", apiPort, "prometheusPort", cfg.ApiConfig.Prometheus.Ports)

	// Run server
	netDone := make(chan struct{}, 1)
	var netErr error
	go func() {
		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			netErr = errs.WrapMsg(err, fmt.Sprintf("api start err: %s", server.Addr))
			netDone <- struct{}{}
		}
	}()
	if cfg.Discovery.Enable == kdisc.ETCDCONST {
		cm := disetcd.NewConfigManager(client.(*etcd.SvcDiscoveryRegistryImpl).GetClient(),
			[]string{
				config.WalletAPIWalletCfgFileName,
				config.DiscoveryConfigFileName,
				config.ShareFileName,
				config.LogConfigFileName,
			},
		)
		cm.Watch(ctx)
	}

	timeoutShutdown := 15 * time.Second
	return gracefulShutdown(server, timeoutShutdown, netDone, netErr)
}

func shutdown(server *http.Server, timeout time.Duration) func() error {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return errs.WrapMsg(err, "shutdown err")
		}
		return nil
	}
}

func gracefulShutdown(server *http.Server, timeout time.Duration, netDone chan struct{}, netErr error) error {
	// register shutdown hook
	sd := shutdown(server, timeout)
	disetcd.RegisterShutDown(sd)

	// handle OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigs)

	select {
	case <-sigs:
		log.CInfo(context.Background(), "received shutdown signal, stopping server...")
		program.SIGTERMExit()
		return sd()

	case <-netDone:
		close(netDone)
		return netErr
	}
}
