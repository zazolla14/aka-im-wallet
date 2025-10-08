package cache

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-tools/utils/stringutil"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/tokenverify"
)

const (
	//nolint:gosec // not a secret, just a cache key prefix
	CacheKeyChatTokenStatus = "CHAT_UID_TOKEN_STATUS:"
	userMaxTokenNum         = 10
)

type TokenInterface interface {
	GetTokensWithoutError(ctx context.Context, userID string) (map[string]int32, error)
}
type TokenCacheRedis struct {
	token *tokenverify.Token
	rdb   redis.UniversalClient
}

func NewTokenInterface(rdb redis.UniversalClient, token *tokenverify.Token) *TokenCacheRedis {
	return &TokenCacheRedis{rdb: rdb, token: token}
}

func (t *TokenCacheRedis) GetTokensWithoutError(ctx context.Context, userID string) (map[string]int32, error) {
	key := CacheKeyChatTokenStatus + userID
	m, err := t.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, errs.Wrap(err)
	}
	mm := make(map[string]int32)
	for k, v := range m {
		mm[k] = stringutil.StringToInt32(v)
	}
	return mm, nil
}
