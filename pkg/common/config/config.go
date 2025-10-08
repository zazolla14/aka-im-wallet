package config

import (
	_ "embed"
	"strconv"

	"github.com/1nterdigital/aka-im-tools/db/mysqlutil"
	"github.com/1nterdigital/aka-im-tools/db/pgutil"
	"github.com/1nterdigital/aka-im-tools/db/redisutil"
	"github.com/1nterdigital/aka-im-tools/xtls"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/kafka"
)

var (
	//go:embed version
	Version string
)

type Log struct {
	StorageLocation     string `mapstructure:"storageLocation"`
	RotationTime        uint   `mapstructure:"rotationTime"`
	RemainRotationCount uint   `mapstructure:"remainRotationCount"`
	RemainLogLevel      int    `mapstructure:"remainLogLevel"`
	IsStdout            bool   `mapstructure:"isStdout"`
	IsJson              bool   `mapstructure:"isJson"`
	IsSimplify          bool   `mapstructure:"isSimplify"`
	WithStack           bool   `mapstructure:"withStack"`
}

type API struct {
	Api struct {
		ListenIP string `mapstructure:"listenIP"`
		Ports    []int  `mapstructure:"ports"`
	} `mapstructure:"api"`
	Prometheus struct {
		Enable       bool   `mapstructure:"enable"`
		AutoSetPorts bool   `mapstructure:"autoSetPorts"`
		Ports        []int  `mapstructure:"ports"`
		GrafanaURL   string `mapstructure:"grafanaURL"`
	} `mapstructure:"prometheus"`
}

type Discovery struct {
	Enable     string     `mapstructure:"enable"`
	Etcd       Etcd       `mapstructure:"etcd"`
	Kubernetes Kubernetes `mapstructure:"kubernetes"`
}

type RedisTLSConfig struct {
	EnableTLS  bool   `mapstructure:"enableTLS"`
	ServerName string `mapstructure:"serverName"`
}

type Redis struct {
	Address     []string       `mapstructure:"address"`
	Username    string         `mapstructure:"username"`
	Password    string         `mapstructure:"password"`
	ClusterMode bool           `mapstructure:"clusterMode"`
	DB          int            `mapstructure:"storage"`
	MaxRetry    int            `mapstructure:"maxRetry"`
	PoolSize    int            `mapstructure:"poolSize"`
	TLS         RedisTLSConfig `mapstructure:"tls"`
}

func (r *Redis) Build() *redisutil.Config {
	conf := &redisutil.Config{
		ClusterMode: r.ClusterMode,
		Address:     r.Address,
		Username:    r.Username,
		Password:    r.Password,
		DB:          r.DB,
		MaxRetry:    r.MaxRetry,
		PoolSize:    r.PoolSize,
	}

	if r.TLS.EnableTLS {
		conf.TLS = &xtls.ClientConfig{
			ServerName: r.TLS.ServerName,
		}
	}

	return conf
}

type Postgres struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Database       string `mapstructure:"database"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	SSLMode        string `mapstructure:"sslmode"`
	MaxOpenConns   int    `mapstructure:"maxOpenConns"`
	MaxIdleConns   int    `mapstructure:"maxIdleConns"`
	MaxRetry       int    `mapstructure:"maxRetry"`
	ConnectTimeout int    `mapstructure:"connectTimeout"`
}

func (p *Postgres) Build() *pgutil.Config {
	return &pgutil.Config{
		Host:           p.Host,
		Port:           p.Port,
		Database:       p.Database,
		Username:       p.Username,
		Password:       p.Password,
		SSLMode:        p.SSLMode,
		MaxOpenConns:   p.MaxOpenConns,
		MaxIdleConns:   p.MaxIdleConns,
		MaxRetry:       p.MaxRetry,
		ConnectTimeout: p.ConnectTimeout,
	}
}

type Mysql struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Database       string `mapstructure:"database"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	MaxOpenConns   int    `mapstructure:"maxOpenConns"`
	MaxIdleConns   int    `mapstructure:"maxIdleConns"`
	MaxRetry       int    `mapstructure:"maxRetry"`
	ConnectTimeout int    `mapstructure:"connectTimeout"`
	URI            string `mapstructure:"uri"` // Optional override
}

func (m *Mysql) Build() *mysqlutil.Config {
	return &mysqlutil.Config{
		Host:         m.Host,
		Port:         strconv.Itoa(m.Port),
		Database:     m.Database,
		Username:     m.Username,
		Password:     m.Password,
		MaxOpenConns: m.MaxOpenConns,
		MaxIdleConns: m.MaxIdleConns,
		MaxRetry:     m.MaxRetry,
		URI:          m.URI,
	}
}

type Kubernetes struct {
	Namespace string `mapstructure:"namespace"`
}

type Etcd struct {
	RootDirectory string   `mapstructure:"rootDirectory"`
	Address       []string `mapstructure:"address"`
	Username      string   `mapstructure:"username"`
	Password      string   `mapstructure:"password"`
}

type Share struct {
	AkaIM struct {
		ApiURL      string `mapstructure:"apiURL"`
		Secret      string `mapstructure:"secret"`
		AdminUserID string `mapstructure:"adminUserID"`
	} `mapstructure:"AkaIM"`
	WalletAdmin []string `mapstructure:"walletAdmin"`
	ProxyHeader string   `mapstructure:"proxyHeader"`
	DBOption    string   `mapstructure:"dbOption"`
}
type Admin struct {
	TokenPolicy struct {
		Expire int `mapstructure:"expire"`
	} `mapstructure:"tokenPolicy"`
	Secret string `mapstructure:"secret"`
}

type MsgTransfer struct {
	Prometheus struct {
		Enable bool  `mapstructure:"enable"`
		Ports  []int `mapstructure:"ports"`
	} `mapstructure:"prometheus"`
}

type Publisher struct {
	Prometheus struct {
		Enable bool  `mapstructure:"enable"`
		Ports  []int `mapstructure:"ports"`
	} `mapstructure:"prometheus"`
}

type Kafka struct {
	Username                 string   `mapstructure:"username"`
	Password                 string   `mapstructure:"password"`
	ProducerAck              string   `mapstructure:"producerAck"`
	CompressType             string   `mapstructure:"compressType"`
	Address                  []string `mapstructure:"address"`
	ToExpiredTransferTopic   string   `mapstructure:"toExpiredTransferTopic"`
	ToExpiredEnvelopeTopic   string   `mapstructure:"toExpiredEnvelopeTopic"`
	ToExpiredTransferGroupID string   `mapstructure:"toExpiredTransferGroupID"`
	ToExpiredEnvelopeGroupID string   `mapstructure:"toExpiredEnvelopeGroupID"`

	Tls TLSConfig `mapstructure:"tls"`
}

type TLSConfig struct {
	EnableTLS          bool   `mapstructure:"enableTLS"`
	CACrt              string `mapstructure:"caCrt"`
	ClientCrt          string `mapstructure:"clientCrt"`
	ClientKey          string `mapstructure:"clientKey"`
	ClientKeyPwd       string `mapstructure:"clientKeyPwd"`
	InsecureSkipVerify bool   `mapstructure:"insecureSkipVerify"`
}

type Tracer struct {
	AppName struct {
		Api string `mapstructure:"api"`
	} `mapstructure:"appName"`
	Otel struct {
		Collector struct {
			Address string `mapstructure:"address"`
		} `mapstructure:"collector"`
	} `mapstructure:"otel"`

	Enable bool `mapstructure:"enable"`
}

func (k *Kafka) Build() *kafka.Config {
	return &kafka.Config{
		Username:     k.Username,
		Password:     k.Password,
		ProducerAck:  k.ProducerAck,
		CompressType: k.CompressType,
		Addr:         k.Address,
		TLS: kafka.TLSConfig{
			EnableTLS:          k.Tls.EnableTLS,
			CACrt:              k.Tls.CACrt,
			ClientCrt:          k.Tls.ClientCrt,
			ClientKey:          k.Tls.ClientKey,
			ClientKeyPwd:       k.Tls.ClientKeyPwd,
			InsecureSkipVerify: k.Tls.InsecureSkipVerify,
		},
	}
}
