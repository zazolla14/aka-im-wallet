package wallet

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type repositoryImpl struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) FindByUserID(userID string) (resp *entity.Wallet, err error) {
	var wallet entity.Wallet
	if err := r.db.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (r *repositoryImpl) Create(wallet *entity.Wallet) (err error) {
	if err := r.db.Create(wallet).Error; err != nil {
		return err
	}
	return nil
}

func (r *repositoryImpl) FindByWalletID(walletID int64) (resp *entity.Wallet, err error) {
	var wallet entity.Wallet
	if err := r.db.Where("wallet_id = ?", walletID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

//nolint:revive // keep receiver for interface compliance
func (r *repositoryImpl) GetWalletByWalletIDTx(
	ctx context.Context, tx *gorm.DB, walletID int64,
) (wallet *entity.Wallet, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	err = tx.WithContext(ctx).Clauses(
		clause.Locking{
			Strength: "UPDATE",
		},
	).
		Where("wallet_id = ? AND is_active IS TRUE", walletID).
		First(&wallet, walletID).Error
	if err != nil {
		log.ZError(ctx, "while Locking wallet", err)
		return wallet, err
	}

	span.SetAttributes(attribute.Int64("walletID", walletID))

	return wallet, nil
}

func (r *repositoryImpl) UpdateWallet(ctx context.Context, tx *gorm.DB, wallet *entity.Wallet) error {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
		err      error
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	span.SetAttributes(attribute.Int64("walletID", wallet.WalletID))

	switch tx {
	case nil:
		err = r.db.WithContext(ctx).Model(&entity.Wallet{}).
			Where("wallet_id = ? AND is_active IS TRUE", wallet.WalletID).
			Updates(wallet).Error
	default:
		err = tx.WithContext(ctx).Model(&entity.Wallet{}).
			Where("wallet_id = ? AND is_active IS TRUE", wallet.WalletID).
			Updates(wallet).Error
	}

	if err != nil {
		log.ZError(ctx, "while repositoryImpl UpdateWallet", err)
		return err
	}

	return nil
}

func (r *repositoryImpl) GetWalletByUserID(
	ctx context.Context, userID string,
) (wallet *entity.Wallet, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	err = r.db.WithContext(ctx).
		Where("user_id = ? AND is_active IS TRUE", userID).
		First(&wallet).Error
	if err != nil {
		log.ZError(ctx, "while repositoryImpl GetWalletByUserID", err)
		return wallet, err
	}

	span.SetAttributes(
		attribute.String("userID", userID),
		attribute.Int64("walletID", wallet.WalletID),
	)

	return wallet, nil
}

func (r *repositoryImpl) CreateWallet(
	ctx context.Context, wallet *entity.Wallet,
) (walletID int64, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	err = r.db.WithContext(ctx).Create(wallet).Error
	if err != nil {
		log.ZError(ctx, "while create wallet", err)
		return walletID, err
	}
	span.SetAttributes(
		attribute.Int64("walletID", wallet.WalletID),
	)

	return wallet.WalletID, nil
}
