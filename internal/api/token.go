package api

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

func (o walletService) ParseToken(ctx context.Context, req *domain.ParseTokenRequest) (*domain.ParseTokenResponse, error) {
	userID, userType, err := o.Token.GetToken(req.Token)
	if err != nil {
		return nil, err
	}
	m, err := o.Database.GetTokens(ctx, userID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if len(m) == 0 {
		return nil, eerrs.ErrTokenNotExist.Wrap()
	}
	if _, ok := m[req.Token]; !ok {
		return nil, eerrs.ErrTokenNotExist.Wrap()
	}

	return &domain.ParseTokenResponse{
		UserID:   userID,
		UserType: userType,
	}, nil
}
