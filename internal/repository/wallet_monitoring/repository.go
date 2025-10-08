//go:generate mockgen -source=$GOFILE -destination=$PROJECT_DIR/generated/mock/mock_$GOPACKAGE/$GOFILE

package wallet_monitoring

import (
	"context"

	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type WalletMonitoringRepository interface {
	GetDashboardTransactionVolume(ctx context.Context) ([]*entity.TransactionCount, error)
	GetListTransactionMonitoring(
		ctx context.Context, req *domain.GetListTransactionMonitoringRequest,
	) ([]*entity.WalletTransaction, int64, error)
	GetListEnvelope(
		ctx context.Context, req *domain.GetListEnvelopeRequest,
	) ([]*entity.Envelope, int64, error)
	GetEnvelopeDetail(
		ctx context.Context, req *domain.GetEnvelopeDetailRequest,
	) (*entity.Envelope, []*entity.EnvelopeDetail, error)
	GetTransferHistory(
		ctx context.Context, req *domain.GetTransferHistoryRequest,
	) ([]*entity.Transfer, int64, error)
	GetTop10Users(
		ctx context.Context, req *domain.GetTop10UsersRequest,
	) ([]*entity.UserStatResult, string, error)
}
