package identity

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

const expiryBatchSize = 100

type Expiry struct{ repository *Repository }

func NewExpiry(repository *Repository) *Expiry { return &Expiry{repository: repository} }

// RemoveExpired requires a tenant context. Ending ctx stops it between batches and leaves the remaining rows for the
// next call; a lock held by a concurrent request also defers that row to the next call.
func (e *Expiry) RemoveExpired(ctx context.Context) error {
	now := timex.Now()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		sessions, recoveryTokens, err := e.repository.deleteExpiredCredentials(ctx, now, expiryBatchSize)
		if err != nil {
			return err
		}
		if sessions < expiryBatchSize && recoveryTokens < expiryBatchSize {
			return nil
		}
	}
}
