//go:generate mockgen -source=$GOFILE -destination=$PROJECT_DIR/generated/mock/mock_$GOPACKAGE/$GOFILE

package envelope

import (
	"context"
	"time"

	"gorm.io/gorm"

	e "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type Repository interface {
	GetAllEnvelopesByUserID(userID int64) (resp []*e.Envelope, err error)
	GetEnvelope(ctx context.Context, envelopeID int64) (resp *e.Envelope, err error)
	CreateEnvelope(ctx context.Context, envelope *e.Envelope) (err error)
	CountSentEnvelope(ctx context.Context, userID string, now time.Time) (count int64, err error)
	CountClaimedEnvelope(ctx context.Context, userID string, now time.Time) (count int64, err error)
	LockEnvelopeByID(ctx context.Context, envelopeID int64) (resp *e.Envelope, err error)
	ClaimNextLuckyShare(ctx context.Context, envelopeID int64) (resp *e.EnvelopeDetail, err error)
	UpdateEnvelopeDetail(ctx context.Context, detail *e.EnvelopeDetail, tx *gorm.DB) (err error)
	UpdateClaimedAmount(ctx context.Context, envelopeID int64, amount float64) (err error)
	GetExpiredUnRefundedEnvelopes(ctx context.Context) (resp []*e.Envelope, err error)
	GetExpiredUnRefundedEnvelopesByID(ctx context.Context, envelopeID int64, userID string) (resp *e.Envelope, err error)
	RefundEnvelope(ctx context.Context, envelopeID int64, refundAmount float64) (err error)
	DeactivateUnclaimedDetails(ctx context.Context, envelopeID int64) (err error)
	CreateEnvelopeDetails(ctx context.Context, envelopeDetail []*e.EnvelopeDetail) (err error)
	CheckClaimStatus(ctx context.Context, envelopeID int64, userID string) (resp *e.ClaimStatus, err error)
	WithTransaction(ctx context.Context, fn func(txRepo Repository) error) (err error)
	WithTx(tx *gorm.DB) Repository
	GetTx() *gorm.DB
	Update(ctx context.Context, envelope *e.Envelope, tx *gorm.DB) (err error)
	GetEnvelopeDetail(ctx context.Context, envelopeDetailID int64) (resp *e.EnvelopeDetail, err error)
	GetEnvelopeDetailsByEnvelopID(ctx context.Context, envelopID int64) (resp []*e.EnvelopeDetail, err error)
	FetchExpiredEnvelopes(ctx context.Context, ids []int64) (resp []*e.Envelope, err error)
}
