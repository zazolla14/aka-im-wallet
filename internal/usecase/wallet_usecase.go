package usecase

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet_transaction"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

type (
	WalletSvcImpl struct {
		repo                        wallet.Repository
		walletTransactionRepository wallet_transaction.Repository
		trx                         *gorm.DB
	}

	WalletSvc interface {
		GetWalletByUserID(userID string) (resp *domain.WalletDomain, err error)
		GetWalletDetail(ctx context.Context, userID string) (wallet *domain.Wallet, err error)
		CreateWallet(
			ctx context.Context, userID, createdBy string,
		) (wallet *domain.Wallet, err error)
	}
)

func NewWalletUseCase(
	repo wallet.Repository,
	trx *gorm.DB,
	walletTransactionRepo wallet_transaction.Repository,
) WalletSvc {
	return &WalletSvcImpl{
		repo:                        repo,
		trx:                         trx,
		walletTransactionRepository: walletTransactionRepo,
	}
}

func (s *WalletSvcImpl) GetWalletByUserID(userID string) (resp *domain.WalletDomain, err error) {
	walletDB, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	return toDomainWallet(walletDB), nil
}

func toDomainWallet(walletDB *entity.Wallet) (resp *domain.WalletDomain) {
	return &domain.WalletDomain{
		WalletID:  walletDB.WalletID,
		Balance:   walletDB.Balance,
		IsActive:  walletDB.IsActive,
		CreatedAt: walletDB.CreatedAt,
	}
}

func (u *WalletSvcImpl) GetWalletDetail(
	ctx context.Context, userID string,
) (walletDetail *domain.Wallet, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelUsecase)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	var resp *entity.Wallet
	resp, err = u.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.ZError(ctx, "while repo.GetWalletByUserID", err, "userID", userID)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("userIDReq", userID),
		attribute.Int64("walletIDResp", resp.WalletID),
	)

	return dtoWalletDetail(resp), nil
}

func (u *WalletSvcImpl) CreateWallet(
	ctx context.Context, userID string, createdBy string,
) (resp *domain.Wallet, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelUsecase)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	var walletUser *entity.Wallet
	walletUser, err = u.repo.GetWalletByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.ZError(ctx, "while repo.GetWalletByUserID", err, "userID", userID)
		return nil, err
	}

	if walletUser != nil && walletUser.WalletID > 0 {
		log.ZError(ctx, "validate wallet existed", err, "userID", userID)
		return nil, eerrs.ErrWalletExisted
	}

	var walletID int64
	walletID, err = u.repo.CreateWallet(ctx, &entity.Wallet{
		UserID:    userID,
		IsActive:  true,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	})
	if err != nil {
		log.ZError(ctx, "while repo.CreateWallet", err, "userID", userID)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("userIDReq", userID),
		attribute.Int64("walletIDResp", walletID),
		attribute.String("createdBy", createdBy),
	)

	return &domain.Wallet{
		ID:        walletID,
		UserID:    userID,
		CreatedBy: createdBy,
	}, nil
}
