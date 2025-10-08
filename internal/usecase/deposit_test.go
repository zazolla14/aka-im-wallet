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

	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_tx"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_usecase"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_wallet"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_wallet_recharge_request"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

func Test_ProcessDepositByAdmin(t *testing.T) {
	type expected struct {
		resp *domain.ProcessDepositByAdminResponse
		err  error
	}

	testCases := []struct {
		desc                           string
		req                            *domain.ProcessDepositByAdminRequest
		expected                       expected
		wantError                      bool
		onMockWalletTransactionUsecase func(mock *mock_usecase.MockWalletTransactionSvc)
		onMockDepositRepo              func(mock *mock_wallet_recharge_request.MockRepository)
		onMockTxRepo                   func(mock *mock_tx.MockRepository)
		onMockWalletRepo               func(mock *mock_wallet.MockRepository)
	}{
		{
			desc:      "WalletNotFound",
			wantError: true,
			req: &domain.ProcessDepositByAdminRequest{
				Amount:      1000.0,
				UserID:      "000000",
				Description: "TICKET01",
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("wallet not found"))
			},
			expected: expected{
				resp: nil,
				err:  errors.New("wallet not found"),
			},
		},
		{
			desc:      "ErrWhileCreatingDepositRepo",
			wantError: true,
			req: &domain.ProcessDepositByAdminRequest{
				Amount:      1000.0,
				UserID:      "123456",
				Description: "TICKET01",
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  1000.0,
						UserID:   "123456",
					}, nil)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockDepositRepo: func(mock *mock_wallet_recharge_request.MockRepository) {
				mock.EXPECT().
					CreateDeposit(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("something went wrong"))
			},
			expected: expected{
				resp: nil,
				err:  errors.New("something went wrong"),
			},
		},
		{
			desc:      "ErrWhileCreatingWalletTransaction",
			wantError: true,
			req: &domain.ProcessDepositByAdminRequest{
				Amount:      1000.0,
				UserID:      "123456",
				Description: "TICKET01",
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  1000.0,
						UserID:   "123456",
					}, nil)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockDepositRepo: func(mock *mock_wallet_recharge_request.MockRepository) {
				mock.EXPECT().
					CreateDeposit(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			onMockWalletTransactionUsecase: func(mock *mock_usecase.MockWalletTransactionSvc) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("something went wrong"))
			},
			expected: expected{
				resp: nil,
				err:  errors.New("something went wrong"),
			},
		},
		{
			desc:      "SuccessProcessDepositByAdmin",
			wantError: false,
			req: &domain.ProcessDepositByAdminRequest{
				Amount:      1000.0,
				UserID:      "123456",
				Description: "TICKET01",
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  1000.0,
						UserID:   "123456",
					}, nil)
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockDepositRepo: func(mock *mock_wallet_recharge_request.MockRepository) {
				mock.EXPECT().
					CreateDeposit(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			onMockWalletTransactionUsecase: func(mock *mock_usecase.MockWalletTransactionSvc) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			expected: expected{
				resp: &domain.ProcessDepositByAdminResponse{
					WalletRechargeRequestID: 1,
				},
				err: nil,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockDepositRepo := mock_wallet_recharge_request.NewMockRepository(ctrl)
			if tC.onMockDepositRepo != nil {
				tC.onMockDepositRepo(onMockDepositRepo)
			}

			onMockTxRepo := mock_tx.NewMockRepository(ctrl)
			if tC.onMockTxRepo != nil {
				tC.onMockTxRepo(onMockTxRepo)
			}

			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			onMockWalletTransactionUsecase := mock_usecase.NewMockWalletTransactionSvc(ctrl)
			if tC.onMockWalletTransactionUsecase != nil {
				tC.onMockWalletTransactionUsecase(onMockWalletTransactionUsecase)
			}

			svc := NewWalletRechargeRequestUseCase(
				nil,
				onMockWalletTransactionUsecase,
				onMockDepositRepo,
				nil,
				onMockWalletRepo,
				onMockTxRepo,
			)

			got, err := svc.ProcessDepositByAdmin(context.Background(), tC.req)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, got.WalletRechargeRequestID, tC.expected.resp.WalletRechargeRequestID)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}

func Test_GetListDeposit(t *testing.T) {
	type expected struct {
		resp *domain.GetListDepositResponse
		err  error
	}

	var now = time.Now()

	testCases := []struct {
		desc              string
		req               *domain.GetListDepositRequest
		expected          expected
		wantError         bool
		onMockDepositRepo func(mock *mock_wallet_recharge_request.MockRepository)
		onMockWalletRepo  func(mock *mock_wallet.MockRepository)
	}{
		{
			desc:      "ErrWalletNotFound",
			wantError: true,
			req: &domain.GetListDepositRequest{
				UserID: "123456",
				Page:   1,
				Limit:  10,
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, eerrs.ErrWalletNotFound)
			},
			expected: expected{
				resp: nil,
				err:  eerrs.ErrWalletNotFound,
			},
		},
		{
			desc:      "ErrInvalidStatusRequest",
			wantError: true,
			req: &domain.GetListDepositRequest{
				StatusRequest: "invalid",
				Page:          1,
				Limit:         10,
			},
			expected: expected{
				resp: nil,
				err:  eerrs.ErrInvalidStatusRequest,
			},
		},
		{
			desc:      "ErrWhileGetListDeposit",
			wantError: true,
			req: &domain.GetListDepositRequest{
				Page:  1,
				Limit: 10,
			},
			onMockDepositRepo: func(mock *mock_wallet_recharge_request.MockRepository) {
				mock.EXPECT().
					GetListDeposit(gomock.Any(), gomock.Any()).
					Return(
						nil,                                // list deposit
						int64(0),                           // total count
						errors.New("something went wrong"), // error
					)
			},
			expected: expected{
				resp: nil,
				err:  errors.New("something went wrong"),
			},
		},
		{
			desc:      "SuccessGetListDeposit",
			wantError: false,
			req: &domain.GetListDepositRequest{
				Page:  1,
				Limit: 10,
			},
			onMockDepositRepo: func(mock *mock_wallet_recharge_request.MockRepository) {
				mock.EXPECT().
					GetListDeposit(gomock.Any(), gomock.Any()).
					Return(
						[]*entity.WalletRechargeRequest{
							{
								WalletRechargeRequestID: 1,
								WalletID:                1,
								Amount:                  1000.0,
								CreatedAt:               now,
							},
							{
								WalletRechargeRequestID: 2,
								WalletID:                1,
								Amount:                  1000.0,
								CreatedAt:               now,
							},
						}, // list deposit
						int64(2), // total count
						nil,      // error
					)
			},
			expected: expected{
				resp: &domain.GetListDepositResponse{
					TotalCount: 2,
					Page:       1,
					Limit:      10,
					Deposits: []*domain.Deposit{
						{
							DepositID: 1,
							WalletID:  1,
							Amount:    1000.0,
							CreatedAt: now,
						},
						{
							DepositID: 2,
							WalletID:  1,
							Amount:    1000.0,
							CreatedAt: now,
						},
					},
				},
				err: nil,
			},
		},
		{
			desc:      "SuccessGetListDepositByUserID",
			wantError: false,
			req: &domain.GetListDepositRequest{
				UserID: "123456",
				Page:   1,
				Limit:  10,
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  1000.0,
						UserID:   "123456",
					}, nil)
			},
			onMockDepositRepo: func(mock *mock_wallet_recharge_request.MockRepository) {
				mock.EXPECT().
					GetListDeposit(gomock.Any(), gomock.Any()).
					Return(
						[]*entity.WalletRechargeRequest{
							{
								WalletRechargeRequestID: 1,
								WalletID:                1,
								Amount:                  1000.0,
								CreatedAt:               now,
							},
							{
								WalletRechargeRequestID: 2,
								WalletID:                1,
								Amount:                  1000.0,
								CreatedAt:               now,
							},
						}, // list deposit
						int64(2), // total count
						nil,      // error
					)
			},
			expected: expected{
				resp: &domain.GetListDepositResponse{
					TotalCount: 2,
					Page:       1,
					Limit:      10,
					Deposits: []*domain.Deposit{
						{
							DepositID: 1,
							WalletID:  1,
							Amount:    1000.0,
							CreatedAt: now,
						},
						{
							DepositID: 2,
							WalletID:  1,
							Amount:    1000.0,
							CreatedAt: now,
						},
					},
				},
				err: nil,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockDepositRepo := mock_wallet_recharge_request.NewMockRepository(ctrl)
			if tC.onMockDepositRepo != nil {
				tC.onMockDepositRepo(onMockDepositRepo)
			}

			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			svc := NewWalletRechargeRequestUseCase(
				nil,
				nil,
				onMockDepositRepo,
				nil,
				onMockWalletRepo,
				nil,
			)

			got, err := svc.GetListDeposit(context.Background(), tC.req)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, got.Limit, tC.expected.resp.Limit)
				assert.Equal(t, got.Page, tC.expected.resp.Page)
				assert.Equal(t, got.TotalCount, tC.expected.resp.TotalCount)
				assert.ElementsMatch(t, got.Deposits, tC.expected.resp.Deposits)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}
