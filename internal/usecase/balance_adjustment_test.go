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

	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_balance_adjustment"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_tx"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_usecase"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_wallet"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

func Test_BalanceAdjustment(t *testing.T) {
	type expected struct {
		resp *domain.BalanceAdjustmentByAdminResponse
		err  error
	}

	testCases := []struct {
		desc                     string
		req                      *domain.BalanceAdjustmentByAdminRequest
		expected                 expected
		wantError                bool
		onMockTransactionUsecase func(mock *mock_usecase.MockWalletTransactionSvc)
		onMockAdjustmentRepo     func(mock *mock_balance_adjustment.MockRepository)
		onMockTxRepo             func(mock *mock_tx.MockRepository)
		onMockWalletRepo         func(mock *mock_wallet.MockRepository)
	}{
		{
			desc: "WalletNotFound",
			req: &domain.BalanceAdjustmentByAdminRequest{
				UserID:      "123456",
				Amount:      100.0,
				Reason:      "kekurangan topup",
				Description: "",
				OperatedBy:  "admin",
			},
			expected: expected{
				resp: &domain.BalanceAdjustmentByAdminResponse{},
				err:  errors.New("wallet not found"),
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("wallet not found"))
			},
		},
		{
			desc: "InsufficientBalanceUser",
			req: &domain.BalanceAdjustmentByAdminRequest{
				UserID:      "123456",
				Amount:      -100.0,
				Reason:      "kelebihan topup",
				Description: "",
				OperatedBy:  "admin",
			},
			expected: expected{
				resp: &domain.BalanceAdjustmentByAdminResponse{},
				err:  errors.New("insufficient balance user 123456"),
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "123456",
						Balance:  0,
					}, nil)
			},
		},
		{
			desc: "ErrCreateDataBalanceAdjustment",
			req: &domain.BalanceAdjustmentByAdminRequest{
				UserID:      "123456",
				Amount:      100.0,
				Reason:      "kekurangan topup",
				Description: "",
				OperatedBy:  "admin",
			},
			expected: expected{
				resp: &domain.BalanceAdjustmentByAdminResponse{},
				err:  errors.New("something went wrong"),
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "123456",
						Balance:  0,
					}, nil)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockAdjustmentRepo: func(mock *mock_balance_adjustment.MockRepository) {
				mock.EXPECT().
					CreateBalanceAdjustment(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("something went wrong"))
			},
		},
		{
			desc: "ErrCreateTransactionAdjustment",
			req: &domain.BalanceAdjustmentByAdminRequest{
				UserID:      "123456",
				Amount:      100.0,
				Reason:      "kekurangan topup",
				Description: "",
				OperatedBy:  "admin",
			},
			expected: expected{
				resp: &domain.BalanceAdjustmentByAdminResponse{},
				err:  errors.New("something went wrong"),
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "123456",
						Balance:  0,
					}, nil)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockAdjustmentRepo: func(mock *mock_balance_adjustment.MockRepository) {
				mock.EXPECT().
					CreateBalanceAdjustment(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			onMockTransactionUsecase: func(mock *mock_usecase.MockWalletTransactionSvc) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any()).
					Return(int64(1), errors.New("something went wrong"))
			},
		},
		{
			desc: "SuccessAdjustment_AddedBalance",
			req: &domain.BalanceAdjustmentByAdminRequest{
				UserID:      "123456",
				Amount:      100.0,
				Reason:      "kekurangan topup",
				Description: "",
				OperatedBy:  "admin",
			},
			expected: expected{
				resp: &domain.BalanceAdjustmentByAdminResponse{
					BalanceAdjustmentID: 1,
				},
				err: nil,
			},
			wantError: false,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "123456",
						Balance:  0,
					}, nil)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockAdjustmentRepo: func(mock *mock_balance_adjustment.MockRepository) {
				mock.EXPECT().
					CreateBalanceAdjustment(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			onMockTransactionUsecase: func(mock *mock_usecase.MockWalletTransactionSvc) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
		},
		{
			desc: "SuccessAdjustment_DeductBalance",
			req: &domain.BalanceAdjustmentByAdminRequest{
				UserID:      "123456",
				Amount:      100.0,
				Reason:      "kelebihan topup",
				Description: "",
				OperatedBy:  "admin",
			},
			expected: expected{
				resp: &domain.BalanceAdjustmentByAdminResponse{
					BalanceAdjustmentID: 1,
				},
				err: nil,
			},
			wantError: false,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "123456",
						Balance:  300.0,
					}, nil)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockAdjustmentRepo: func(mock *mock_balance_adjustment.MockRepository) {
				mock.EXPECT().
					CreateBalanceAdjustment(gomock.Any(), gomock.Any(), gomock.Any()).
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

			onMockTransactionUsecase := mock_usecase.NewMockWalletTransactionSvc(ctrl)
			if tC.onMockTransactionUsecase != nil {
				tC.onMockTransactionUsecase(onMockTransactionUsecase)
			}

			onMockTxRepo := mock_tx.NewMockRepository(ctrl)
			if tC.onMockTxRepo != nil {
				tC.onMockTxRepo(onMockTxRepo)
			}

			onMockAdjustmentRepo := mock_balance_adjustment.NewMockRepository(ctrl)
			if tC.onMockAdjustmentRepo != nil {
				tC.onMockAdjustmentRepo(onMockAdjustmentRepo)
			}

			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			svc := NewBalanceAdjustmentUseCase(
				onMockTransactionUsecase,
				onMockAdjustmentRepo,
				onMockWalletRepo,
				onMockTxRepo,
			)

			got, err := svc.BalanceAdjustmentByAdmin(context.Background(), tC.req)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, got.BalanceAdjustmentID, tC.expected.resp.BalanceAdjustmentID)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}

func Test_GetListBalanceAdjustment(t *testing.T) {
	type expected struct {
		resp *domain.GetListBalanceAdjustmentResponse
		err  error
	}

	now := time.Now()

	testCases := []struct {
		desc                 string
		req                  *domain.GetListbalanceAjustmentRequest
		expected             expected
		wantError            bool
		onMockAdjustmentRepo func(mock *mock_balance_adjustment.MockRepository)
		onMockWalletRepo     func(mock *mock_wallet.MockRepository)
	}{
		{
			desc: "WalletNotFoundByUserID",
			req: &domain.GetListbalanceAjustmentRequest{
				Page:   1,
				Limit:  10,
				UserID: "123456",
			},
			expected: expected{
				resp: &domain.GetListBalanceAdjustmentResponse{},
				err:  errors.New("wallet not found"),
			},
			wantError: true,
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("wallet not found"))
			},
		},
		{
			desc: "ErrWhileGetListBalanceAdjustment",
			req: &domain.GetListbalanceAjustmentRequest{
				Page:         1,
				Limit:        10,
				UserID:       "123456",
				FilterDateBy: "salah filter",
				StartDate:    now,
				EndDate:      now.Add(24 * time.Hour),
			},
			expected: expected{
				resp: &domain.GetListBalanceAdjustmentResponse{},
				err:  errors.New("something went wrong"),
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						UserID:   "123456",
						Balance:  300.0,
					}, nil)
			},
			wantError: true,
			onMockAdjustmentRepo: func(mock *mock_balance_adjustment.MockRepository) {
				mock.EXPECT().
					GetListBalanceAdjustment(gomock.Any(), gomock.Any()).
					Return(
						[]*entity.BalanceAdjustments{},
						int64(0),
						errors.New("something went wrong"),
					)
			},
		},
		{
			desc: "SuccessGetListBalanceAdjustment",
			req: &domain.GetListbalanceAjustmentRequest{
				Page:  1,
				Limit: 10,
			},
			expected: expected{
				resp: &domain.GetListBalanceAdjustmentResponse{
					Page:       1,
					Limit:      10,
					TotalCount: 1,
					BalanceAdjustments: []*domain.BalanceAdjustment{
						{
							BalanceAdjustmentID: 1,
							WalletID:            1,
							Amount:              100.0,
							CreatedAt:           now,
							CreatedBy:           "admin",
						},
					},
				},
				err: nil,
			},
			wantError: false,
			onMockAdjustmentRepo: func(mock *mock_balance_adjustment.MockRepository) {
				mock.EXPECT().
					GetListBalanceAdjustment(gomock.Any(), gomock.Any()).
					Return(
						[]*entity.BalanceAdjustments{
							{
								BalanceAdjustmentID: 1,
								WalletID:            1,
								Amount:              100.0,
								CreatedAt:           now,
								CreatedBy:           "admin",
							},
						},
						int64(1),
						nil,
					)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockAdjustmentRepo := mock_balance_adjustment.NewMockRepository(ctrl)
			if tC.onMockAdjustmentRepo != nil {
				tC.onMockAdjustmentRepo(onMockAdjustmentRepo)
			}

			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			svc := NewBalanceAdjustmentUseCase(nil, onMockAdjustmentRepo, onMockWalletRepo, nil)

			got, err := svc.GetListBalanceAdjustment(context.Background(), tC.req)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, got.Page, tC.expected.resp.Page)
				assert.Equal(t, got.Limit, tC.expected.resp.Limit)
				assert.Equal(t, got.TotalCount, tC.expected.resp.TotalCount)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}
