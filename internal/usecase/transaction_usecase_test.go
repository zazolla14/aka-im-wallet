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
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_wallet"
	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_wallet_transaction"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

func Test_CreateTransaction(t *testing.T) {
	type expected struct {
		transactionID int64
		err           error
	}

	testCases := []struct {
		desc                  string
		req                   *domain.CreateTransactionReq
		expected              expected
		wantError             bool
		onMockTransactionRepo func(mock *mock_wallet_transaction.MockRepository)
		onMockTxRepo          func(mock *mock_tx.MockRepository)
		onMockWalletRepo      func(mock *mock_wallet.MockRepository)
	}{
		{
			desc: "ErrWalletIDRequired",
			req: &domain.CreateTransactionReq{
				WalletID:  0,
				Entrytype: "credit",
			},
			expected: expected{
				transactionID: 0,
				err:           eerrs.ErrWalletIDRequired,
			},
			wantError: true,
		},
		{
			desc: "ErrEntryTypeInvalid",
			req: &domain.CreateTransactionReq{
				WalletID:  1,
				Entrytype: "invalid",
			},
			expected: expected{
				transactionID: 0,
				err:           eerrs.ErrEntryTypeInvalid,
			},
			wantError: true,
		},
		{
			desc: "ErrTransactionTypeInvalid",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "credit",
				TransactionType: "invalid",
			},
			expected: expected{
				transactionID: 0,
				err:           eerrs.ErrTransactionTypeInvalid,
			},
			wantError: true,
		},
		{
			desc: "ErrDescriptionEnRequired",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "credit",
				TransactionType: "transfer",
			},
			expected: expected{
				transactionID: 0,
				err:           eerrs.ErrDescriptionEnRequired,
			},
			wantError: true,
		},
		{
			desc: "ErrDescriptionZhRequired",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "credit",
				TransactionType: "transfer",
				DescriptionEn:   "description_en",
			},
			expected: expected{
				transactionID: 0,
				err:           eerrs.ErrDescriptionZhRequired,
			},
			wantError: true,
		},
		{
			desc: "ErrReferenceCodeRequired",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "credit",
				TransactionType: "transfer",
				DescriptionEn:   "description_en",
				DescriptionZh:   "description_zh",
			},
			expected: expected{
				transactionID: 0,
				err:           eerrs.ErrReferenceCodeRequired,
			},
			wantError: true,
		},
		{
			desc: "ErrImpactedItemRequired",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "credit",
				TransactionType: "transfer",
				DescriptionEn:   "description_en",
				DescriptionZh:   "description_zh",
				ReferenceCode:   "reference_code",
			},
			expected: expected{
				transactionID: 0,
				err:           eerrs.ErrImpactedItemRequired,
			},
			wantError: true,
		},
		{
			desc: "error_locking_wallet",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "credit",
				TransactionType: "transfer",
				Amount:          1000.0,
				DescriptionEn:   "description_en",
				DescriptionZh:   "description_zh",
				ReferenceCode:   "reference_code",
				ImpactedItem:    1,
			},
			expected: expected{
				transactionID: 0,
				err:           errors.New("error while locking wallet"),
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByWalletIDTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("error while locking wallet"))
			},
			wantError: true,
		},
		{
			desc: "ErrInsufficientBalance",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "debit",
				TransactionType: "transfer",
				Amount:          1000.0,
				DescriptionEn:   "description_en",
				DescriptionZh:   "description_zh",
				ReferenceCode:   "reference_code",
				ImpactedItem:    1,
			},
			expected: expected{
				transactionID: 0,
				err:           eerrs.ErrInsufficientBalance,
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByWalletIDTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  100.0,
					}, nil)
			},
			wantError: true,
		},
		{
			desc: "error_while_update_wallet",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "debit",
				TransactionType: "transfer",
				Amount:          1000.0,
				DescriptionEn:   "description_en",
				DescriptionZh:   "description_zh",
				ReferenceCode:   "reference_code",
				ImpactedItem:    1,
			},
			expected: expected{
				transactionID: 0,
				err:           errors.New("something went wrong"),
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByWalletIDTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  2000.0,
					}, nil)
				mock.EXPECT().
					UpdateWallet(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("something went wrong"))
			},
			wantError: true,
		},
		{
			desc: "error_while_create_transaction_wallet",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "debit",
				TransactionType: "transfer",
				Amount:          1000.0,
				DescriptionEn:   "description_en",
				DescriptionZh:   "description_zh",
				ReferenceCode:   "reference_code",
				ImpactedItem:    1,
			},
			expected: expected{
				transactionID: 0,
				err:           errors.New("something went wrong"),
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByWalletIDTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  2000.0,
					}, nil)
				mock.EXPECT().
					UpdateWallet(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			onMockTransactionRepo: func(mock *mock_wallet_transaction.MockRepository) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("something went wrong"))
			},
			wantError: true,
		},
		{
			desc: "valid_transaction",
			req: &domain.CreateTransactionReq{
				WalletID:        1,
				Entrytype:       "debit",
				TransactionType: "transfer",
				Amount:          1000.0,
				DescriptionEn:   "description_en",
				DescriptionZh:   "description_zh",
				ReferenceCode:   "reference_code",
				ImpactedItem:    1,
			},
			expected: expected{
				transactionID: 1,
				err:           nil,
			},
			onMockTxRepo: func(mock *mock_tx.MockRepository) {
				mock.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
						return fn(&gorm.DB{})
					})
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByWalletIDTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  2000.0,
					}, nil)
				mock.EXPECT().
					UpdateWallet(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			onMockTransactionRepo: func(mock *mock_wallet_transaction.MockRepository) {
				mock.EXPECT().
					CreateTransaction(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			wantError: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockTransactionRepo := mock_wallet_transaction.NewMockRepository(ctrl)
			if tC.onMockTransactionRepo != nil {
				tC.onMockTransactionRepo(onMockTransactionRepo)
			}
			onMockTxRepo := mock_tx.NewMockRepository(ctrl)
			if tC.onMockTxRepo != nil {
				tC.onMockTxRepo(onMockTxRepo)
			}
			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			svc := NewWalletTransactionUseCase(
				onMockTransactionRepo,
				onMockWalletRepo,
				onMockTxRepo,
			)

			got, err := svc.CreateTransaction(context.Background(), tC.req)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, got, tC.expected.transactionID)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}

func Test_GetListTransaction(t *testing.T) {
	type expected struct {
		resp *domain.GetListTransactionResponse
		err  error
	}

	var now = time.Now()

	testCases := []struct {
		desc                  string
		req                   *domain.GetListTransactionRequest
		expected              expected
		wantError             bool
		onMockTransactionRepo func(mock *mock_wallet_transaction.MockRepository)
		onMockWalletRepo      func(mock *mock_wallet.MockRepository)
	}{
		{
			desc: "error_while_GetWalletByUserID",
			req: &domain.GetListTransactionRequest{
				Page:   1,
				Limit:  10,
				UserID: "1",
			},
			expected: expected{
				resp: nil,
				err:  errors.New("something went wrong"),
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("something went wrong"))
			},
			wantError: true,
		},
		{
			desc: "ErrWalletNotFound",
			req: &domain.GetListTransactionRequest{
				Page:   1,
				Limit:  10,
				UserID: "1",
			},
			expected: expected{
				resp: nil,
				err:  eerrs.ErrWalletNotFound,
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("wallet not found"))
			},
			wantError: true,
		},
		{
			desc: "error_GetListTransaction",
			req: &domain.GetListTransactionRequest{
				Page:   1,
				Limit:  10,
				UserID: "1",
			},
			expected: expected{
				resp: nil,
				err:  errors.New("something went wrong"),
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  2000.0,
					}, nil)
			},
			onMockTransactionRepo: func(mock *mock_wallet_transaction.MockRepository) {
				mock.EXPECT().
					GetListTransaction(gomock.Any(), gomock.Any()).
					Return(nil, int64(0), errors.New("something went wrong"))
			},
			wantError: true,
		},
		{
			desc: "success",
			req: &domain.GetListTransactionRequest{
				Page:   1,
				Limit:  10,
				UserID: "1",
			},
			expected: expected{
				resp: &domain.GetListTransactionResponse{
					Page:       1,
					Limit:      10,
					TotalCount: 1,
					Transactions: []*domain.WalletTransaction{
						{
							WalletTransactionID: 1,
							WalletID:            1,
							TransactionDate:     now,
							TransactionType:     "deposit",
							Amount:              1000.0,
							BeforeBalance:       2000.0,
							AfterBalance:        3000.0,
							EntryType:           "deposit",
							IsShown:             true,
							ImpactedItem:        1,
							DescriptionEn:       "deposit",
							DescriptionZh:       "deposit",
						},
					},
				},
				err: nil,
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(&entity.Wallet{
						WalletID: 1,
						Balance:  3000.0,
					}, nil)
			},
			onMockTransactionRepo: func(mock *mock_wallet_transaction.MockRepository) {
				mock.EXPECT().
					GetListTransaction(gomock.Any(), gomock.Any()).
					Return([]*entity.WalletTransaction{
						{
							WalletTransactionID: 1,
							WalletID:            1,
							TransactionDate:     now,
							TransactionType:     "deposit",
							Amount:              1000.0,
							BeforeBalance:       2000.0,
							AfterBalance:        3000.0,
							EntryType:           "deposit",
							IsShown:             true,
							ImpactedItem:        1,
							DescriptionEN:       "deposit",
							DescriptionZH:       "deposit",
						},
					}, int64(1), nil)
			},
			wantError: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockTransactionRepo := mock_wallet_transaction.NewMockRepository(ctrl)
			if tC.onMockTransactionRepo != nil {
				tC.onMockTransactionRepo(onMockTransactionRepo)
			}

			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			svc := NewWalletTransactionUseCase(
				onMockTransactionRepo,
				onMockWalletRepo,
				nil,
			)

			got, err := svc.GetListTransaction(context.Background(), tC.req)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, got.Limit, tC.expected.resp.Limit)
				assert.Equal(t, got.Page, tC.expected.resp.Page)
				assert.Equal(t, got.TotalCount, tC.expected.resp.TotalCount)
				assert.ElementsMatch(t, got.Transactions, tC.expected.resp.Transactions)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}
