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

	"github.com/1nterdigital/aka-im-wallet/generated/mock/mock_wallet"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

func TestWallet_GetWalletDetail(t *testing.T) {
	type expected struct {
		wallet *domain.Wallet
		err    error
	}

	testCases := []struct {
		desc             string
		userID           string
		expected         expected
		wantError        bool
		onMockWalletRepo func(mock *mock_wallet.MockRepository)
	}{
		{
			desc:   "error_while_GetWalletByUserID",
			userID: "1123453",
			expected: expected{
				wallet: &domain.Wallet{},
				err:    errors.New("something went wrong"),
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("something went wrong"))
			},
			wantError: true,
		},
		{
			desc:   "success_get_existing_wallet",
			userID: "1123453",
			expected: expected{
				wallet: &domain.Wallet{
					ID:        1,
					UserID:    "1123453",
					Balance:   0,
					CreatedAt: time.Now(),
					CreatedBy: "system",
				},
				err: nil,
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(
						&entity.Wallet{
							WalletID:  1,
							UserID:    "1123453",
							Balance:   0,
							CreatedAt: time.Now(),
							CreatedBy: "system",
						}, nil,
					)
			},
			wantError: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			svc := NewWalletUseCase(onMockWalletRepo, nil, nil)

			got, err := svc.GetWalletDetail(context.Background(), tC.userID)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, got.ID, tC.expected.wallet.ID)
				assert.Equal(t, got.UserID, tC.expected.wallet.UserID)
				assert.InDelta(t, got.Balance, tC.expected.wallet.Balance, 0.01)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}

func TestWallet_CreateWallet(t *testing.T) {
	type expected struct {
		wallet *domain.Wallet
		err    error
	}

	testCases := []struct {
		desc             string
		userID           string
		createdBy        string
		expected         expected
		wantError        bool
		onMockWalletRepo func(mock *mock_wallet.MockRepository)
	}{
		{
			desc:      "error_while_GetWalletByUserID",
			userID:    "1123453",
			createdBy: "system",
			expected: expected{
				wallet: &domain.Wallet{},
				err:    errors.New("something went wrong"),
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("something went wrong"))
			},
			wantError: true,
		},
		{
			desc:      "ErrWalletExisted",
			userID:    "1123453",
			createdBy: "system",
			expected: expected{
				wallet: &domain.Wallet{},
				err:    eerrs.ErrWalletExisted,
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(
						&entity.Wallet{
							WalletID:  1,
							UserID:    "1123453",
							Balance:   0,
							CreatedBy: "system",
						}, nil,
					)
			},
			wantError: true,
		},
		{
			desc:      "err_CreateWallet",
			userID:    "1123453",
			createdBy: "system",
			expected: expected{
				wallet: &domain.Wallet{},
				err:    errors.New("something went wrong"),
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, gorm.ErrRecordNotFound)
				mock.EXPECT().
					CreateWallet(gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("something went wrong"))
			},
			wantError: true,
		},
		{
			desc:      "success_create_wallet",
			userID:    "1123453",
			createdBy: "system",
			expected: expected{
				wallet: &domain.Wallet{
					ID:        1,
					UserID:    "1123453",
					Balance:   0,
					CreatedBy: "system",
				},
				err: nil,
			},
			onMockWalletRepo: func(mock *mock_wallet.MockRepository) {
				mock.EXPECT().
					GetWalletByUserID(gomock.Any(), gomock.Any()).
					Return(nil, gorm.ErrRecordNotFound)
				mock.EXPECT().
					CreateWallet(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			wantError: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			onMockWalletRepo := mock_wallet.NewMockRepository(ctrl)
			if tC.onMockWalletRepo != nil {
				tC.onMockWalletRepo(onMockWalletRepo)
			}

			svc := NewWalletUseCase(onMockWalletRepo, nil, nil)

			got, err := svc.CreateWallet(context.Background(), tC.userID, tC.createdBy)
			if !tC.wantError {
				require.NoError(t, err)
				assert.Equal(t, got.ID, tC.expected.wallet.ID)
				assert.Equal(t, got.UserID, tC.expected.wallet.UserID)
				assert.InDelta(t, got.Balance, tC.expected.wallet.Balance, 0.01)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tC.expected.err.Error())
			}
		})
	}
}
