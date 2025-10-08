package wallet_transaction

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type repositoryImpl struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) FindAllByWalletID(walletID int64) ([]entity.WalletTransaction, error) {
	var transactions []entity.WalletTransaction
	if err := r.db.Where("wallet_id = ?", walletID).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *repositoryImpl) FindByWalletTransactionID(walletTransactionID int64) (entity.WalletTransaction, error) {
	var transaction entity.WalletTransaction
	if err := r.db.Where("wallet_transaction_id = ?", walletTransactionID).Find(&transaction).Error; err != nil {
		return entity.WalletTransaction{}, err
	}

	return transaction, nil
}

func (r *repositoryImpl) CreateTransaction(
	ctx context.Context, tx *gorm.DB, transaction *entity.WalletTransaction,
) (transactionID int64, err error) {
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

	switch tx {
	case nil:
		err = r.db.WithContext(ctx).Create(transaction).Error
	default:
		err = tx.WithContext(ctx).Create(transaction).Error
	}

	if err != nil {
		log.ZError(ctx, "while create transaction history", err)
		return transactionID, err
	}

	span.SetAttributes(attribute.Int64("walletTransactionID", transaction.WalletTransactionID))

	return transaction.WalletTransactionID, nil
}

func (r *repositoryImpl) GetListTransaction(
	ctx context.Context, req *domain.GetListTransactionRequest,
) (transactions []*entity.WalletTransaction, total int64, err error) {
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

	span.SetAttributes(
		attribute.Int64("walletID", req.WalletID),
		attribute.String("walletID", req.StartDate.String()),
		attribute.String("walletID", req.EndDate.String()),
	)

	query := r.db.WithContext(ctx).
		Model(&entity.WalletTransaction{}).
		Where("wallet_id = ? AND is_active IS TRUE AND is_shown IS TRUE", req.WalletID)

	if !req.StartDate.IsZero() && !req.EndDate.IsZero() {
		query = query.Where("transaction_date BETWEEN ? AND ?", req.StartDate, req.EndDate)
	}

	err = query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	err = query.Order("transaction_date DESC").
		Limit(int(req.Limit)).
		Offset(int(offset)).
		Find(&transactions).Error
	if err != nil {
		log.ZError(ctx, "while repositoryImpl GetListTransaction", err)
		return nil, 0, err
	}

	return transactions, total, nil
}
