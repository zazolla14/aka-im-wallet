package database

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/1nterdigital/aka-im-tools/db/mysqlutil"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/cache"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/tokenverify"
)

type WalletDatabaseInterface interface {
	GetTokens(ctx context.Context, userID string) (map[string]int32, error)
}

func NewWalletDatabase(
	mysqlCli *mysqlutil.Client,
	rdb redis.UniversalClient,
	token *tokenverify.Token,
) (db WalletDatabaseInterface, err error) {
	return &WalletDatabase{
		mysqlDB: mysqlCli,
		cache:   cache.NewTokenInterface(rdb, token),
	}, nil
}

type WalletDatabase struct {
	cache   cache.TokenInterface
	mysqlDB *mysqlutil.Client
}

func (o *WalletDatabase) GetTokens(ctx context.Context, userID string) (tokens map[string]int32, err error) {
	return o.cache.GetTokensWithoutError(ctx, userID)
}
