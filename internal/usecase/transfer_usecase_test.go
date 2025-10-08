package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_transfer"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_tx"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_usecase"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_wallet"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

func TestTransfer_CreateTransfer(t *testing.T) {
	type expected struct {
		resp *domain.CreateTransferResponse
		err  error
	}
	testCases := []struct {
		desc                     string
		arg                      *domain.CreateTransferRequest
		expected                 expected
		wantError                bool
		onMockTxRepo             func(mock *mock_tx.MockRepository)
		onMockWalletRepo         func(mock *mock_wallet.MockRepository)
		onMockTransferRepo       func(mock *mock_transfer.MockRepository)
		onMockTransactionUsecase func(mock *mock_usecase.MockWalletTransactionSvc)
	}{
		{
			desc: "error_while_GetSourceWallet",
			arg: &domain.CreateTransferRequest{
				FromUserID: "1",
				ToUserID:   "2",
				Amount:     100,
			},
			expected: expected{
				resp: nil,
				err:  errors.New("something wrong"),
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("something wrong"))
			},
		},
		{
			desc: "ErrInsufficientBalance",
			arg: &domain.CreateTransferRequest{
				FromUserID: "1",
				ToUserID:   "2",
				Amount:     100,
			},
			expected: expected{

				resp: nil,
				err:  eerrs.ErrInsufficientBalance,
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "1",
						Balance:  10,
					}, nil).Times(1)
			},
		},
		{
			desc: "error_while_GetTargetWallet",
			arg: &domain.CreateTransferRequest{
				FromUserID: "1",
				ToUserID:   "2",
				Amount:     100,
			},
			expected: expected{

				resp: nil,
				err:  errors.New("something wrong"),
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "1",
						Balance:  1000,
					}, nil).Times(1)
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("something wrong")).Times(1)
			},
		},
		{
			desc: "Err_CreateTransfer",
			arg: &domain.CreateTransferRequest{
				FromUserID: "1",
				ToUserID:   "2",
				Amount:     100,
			},
			expected: expected{

				resp: nil,
				err:  errors.New("something wrong"),
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "1",
						Balance:  1000,
					}, nil).Times(1)
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 2,
						UserID:   "2",
						Balance:  100,
					}, nil).Times(1)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockTransferRepo: func(mock *mock_transfer.MockRepository) {
				mock.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("something wrong")).Times(1)
			},
		},
		{
			desc: "Err_CreateTransaction",
			arg: &domain.CreateTransferRequest{
				FromUserID: "1",
				ToUserID:   "2",
				Amount:     100,
			},
			expected: expected{

				resp: nil,
				err:  errors.New("something wrong"),
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "1",
						Balance:  1000,
					}, nil).Times(1)
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 2,
						UserID:   "2",
						Balance:  100,
					}, nil).Times(1)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockTransferRepo: func(mock *mock_transfer.MockRepository) {
				mock.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			onMockTransactionUsecase: func(mock *mock_usecase.MockWalletTransactionSvc) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("something wrong")).Times(1)
			},
		},
		{
			desc: "SuccessCreateTransfer",
			arg: &domain.CreateTransferRequest{
				FromUserID: "1",
				ToUserID:   "2",
				Amount:     100,
			},
			expected: expected{
				resp: &domain.CreateTransferResponse{
					TransferID: 1,
				},
				err: nil,
			},
			wantError: false,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "1",
						Balance:  1000,
					}, nil).Times(1)
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 2,
						UserID:   "2",
						Balance:  100,
					}, nil).Times(1)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockTransferRepo: func(mock *mock_transfer.MockRepository) {
				mock.EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			onMockTransactionUsecase: func(mock *mock_usecase.MockWalletTransactionSvc) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			onMockTransferRepo := mock_transfer.NewMockRepository(ctrl)
			if tC.onMockTransferRepo != nil {
				tC.onMockTransferRepo(onMockTransferRepo)
			}

			onMockTxRepo := mock_tx.NewMockRepository(ctrl)
			if tC.onMockTxRepo != nil {
				tC.onMockTxRepo(onMockTxRepo)
			}

			onMockTransactionUsecase := mock_usecase.NewMockWalletTransactionSvc(ctrl)
			if tC.onMockTransactionUsecase != nil {
				tC.onMockTransactionUsecase(onMockTransactionUsecase)
			}

			svc := NewTransferUseCase(
				nil,
				onMockTransactionUsecase,
				onMockTransferRepo,
				onMockWalletRepo,
				onMockTxRepo,
			)

			transfer, err := svc.CreateTransfer(context.Background(), tC.arg)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, transfer.TransferID, tC.expected.resp.TransferID)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}

func TestTransfer_RefundTransfer(t *testing.T) {
	testCases := []struct {
		desc                     string
		arg                      *domain.RefundTransferReq
		err                      error
		wantError                bool
		onMockTxRepo             func(mock *mock_tx.MockRepository)
		onMockWalletRepo         func(mock *mock_wallet.MockRepository)
		onMockTransferRepo       func(mock *mock_transfer.MockRepository)
		onMockTransactionUsecase func(mock *mock_usecase.MockWalletTransactionSvc)
	}{
		{
			desc: "ErrWalletNotFound",
			arg: &domain.RefundTransferReq{
				TransferID: 1,
			},
			err:       eerrs.ErrWalletNotFound,
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, eerrs.ErrWalletNotFound)
			},
		},
		{
			desc: "ErrNotEligibleRefundTransfer",
			arg: &domain.RefundTransferReq{
				TransferID: 1,
			},
			err:       errors.New("not eligible refund transfer"),
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "1",
						Balance:  1000,
					}, nil)
			},
			onMockTransferRepo: func(mock *mock_transfer.MockRepository) {
				mock.EXPECT().
					GetEligibleRefundTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						nil,
						errors.New("not eligible refund transfer"),
					)
			},
		},
		{
			desc: "ErrWhileUpdateTransfer",
			arg: &domain.RefundTransferReq{
				TransferID: 1,
			},
			err:       errors.New("while update transfer"),
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "1",
						Balance:  1000,
					}, nil)
			},
			onMockTransferRepo: func(mock *mock_transfer.MockRepository) {
				mock.EXPECT().
					GetEligibleRefundTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						&entity.Transfer{
							TransferID:     1,
							FromUserID:     "111111",
							ToUserID:       "222222",
							Amount:         1000.0,
							StatusTransfer: "pending",
						},
						nil,
					)
				mock.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("while update transfer"))
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
		},
		{
			desc: "ErrWhileCreateTransactions",
			arg: &domain.RefundTransferReq{
				TransferID: 1,
			},
			err:       errors.New("while create transactions"),
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "1",
						Balance:  1000,
					}, nil)
			},
			onMockTransferRepo: func(mock *mock_transfer.MockRepository) {
				mock.EXPECT().
					GetEligibleRefundTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						&entity.Transfer{
							TransferID:     1,
							FromUserID:     "111111",
							ToUserID:       "222222",
							Amount:         1000.0,
							StatusTransfer: "pending",
						},
						nil,
					)
				mock.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockTransactionUsecase: func(mock *mock_usecase.MockWalletTransactionSvc) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("while create transactions"))
			},
		},
		{
			desc: "SuccessCreateTransfer",
			arg: &domain.RefundTransferReq{
				TransferID: 1,
			},
			err:       nil,
			wantError: false,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "1",
						Balance:  1000,
					}, nil)
			},
			onMockTransferRepo: func(mock *mock_transfer.MockRepository) {
				mock.EXPECT().
					GetEligibleRefundTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						&entity.Transfer{
							TransferID:     1,
							FromUserID:     "111111",
							ToUserID:       "222222",
							Amount:         1000.0,
							StatusTransfer: "pending",
						},
						nil,
					)
				mock.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockTransactionUsecase: func(mock *mock_usecase.MockWalletTransactionSvc) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			onMockTransferRepo := mock_transfer.NewMockRepository(ctrl)
			if tC.onMockTransferRepo != nil {
				tC.onMockTransferRepo(onMockTransferRepo)
			}

			onMockTxRepo := mock_tx.NewMockRepository(ctrl)
			if tC.onMockTxRepo != nil {
				tC.onMockTxRepo(onMockTxRepo)
			}

			onMockTransactionUsecase := mock_usecase.NewMockWalletTransactionSvc(ctrl)
			if tC.onMockTransactionUsecase != nil {
				tC.onMockTransactionUsecase(onMockTransactionUsecase)
			}

			svc := NewTransferUseCase(
				nil,
				onMockTransactionUsecase,
				onMockTransferRepo,
				onMockWalletRepo,
				onMockTxRepo,
			)

			err := svc.RefundTransfer(context.Background(), tC.arg)
			if !tC.wantError {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.err.Error())
			}
		})
	}
}

func TestTransfer_GetDetailTransfer(t *testing.T) {
	type (
		expected struct {
			resp *domain.Transfer
			err  error
		}

		arg struct {
			transferID int64
			userID     string
		}
	)

	now := time.Now()

	testCases := []struct {
		desc               string
		arg                arg
		wantError          bool
		expected           expected
		onMockTransferRepo func(mock *mock_transfer.MockRepository)
	}{
		{
			desc: "ErrDetailTransferNotFound",
			arg: arg{
				transferID: 1,
				userID:     "123456",
			},
			wantError: true,
			onMockTransferRepo: func(mock *mock_transfer.MockRepository) {
				mock.EXPECT().
					GetDetailTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						nil,
						errors.New("not eligible refund transfer"),
					)
			},
			expected: expected{
				resp: nil,
				err:  errors.New("not eligible refund transfer"),
			},
		},
		{
			desc: "SuccessGetDetailTransfer",
			arg: arg{
				transferID: 1,
				userID:     "123456",
			},
			wantError: false,
			onMockTransferRepo: func(mock *mock_transfer.MockRepository) {
				mock.EXPECT().
					GetDetailTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						&entity.Transfer{
							TransferID:     1,
							FromUserID:     "111111",
							ToUserID:       "222222",
							Amount:         1000.0,
							StatusTransfer: "pending",
							Remark:         "remark",
							ExpiredAt:      &now,
							RefundedAt:     nil,
							ClaimedAt:      nil,
							CreatedAt:      now,
							CreatedBy:      "SYSTEM",
							UpdatedAt:      now,
							UpdatedBy:      "SYSTEM",
						},
						nil,
					)
			},
			expected: expected{
				resp: &domain.Transfer{
					TransferID:     1,
					FromUserID:     "111111",
					ToUserID:       "222222",
					Amount:         1000.0,
					StatusTransfer: "pending",
					Remark:         "remark",
					ExpiredAt:      &now,
					RefundedAt:     nil,
					ClaimedAt:      nil,
					CreatedAt:      now,
					CreatedBy:      "SYSTEM",
					UpdatedAt:      now,
					UpdatedBy:      "SYSTEM",
				},
				err: nil,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockTransferRepo := mock_transfer.NewMockRepository(ctrl)
			if tC.onMockTransferRepo != nil {
				tC.onMockTransferRepo(onMockTransferRepo)
			}

			svc := NewTransferUseCase(nil, nil, onMockTransferRepo, nil, nil)

			got, err := svc.GetDetailTransfer(context.Background(), tC.arg.transferID, tC.arg.userID)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, got, tC.expected.resp)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}
