package authz

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestRelationshipsWrite(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	pool, statements := newLoggedPool(t, db)
	writer := permissions.NewRelationships(NewRepository(pool), testModel)
	production := tenant.Production().String()
	userID, contestID, groupID, logID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	admin, moderators, link := siteAdminRelationship(userID), contestModeratorGroup(contestID, groupID),
		logCountsFor(logID, contestID)
	owner, invalid := contestOwner(contestID, userID), permissions.Relationship{}
	batch := []permissions.Relationship{admin, moderators, link, admin}
	granted, afterRevoke := rows(batch), rows([]permissions.Relationship{admin, moderators})

	steps := []struct {
		name           string
		write          func() error
		wantErr        bool
		wantRows       []string
		wantStatements int
	}{
		{"grant a batch with a duplicate", func() error { return writer.Grant(ctx, batch...) }, false, granted, 1},
		{"grant the batch again", func() error { return writer.Grant(ctx, batch...) }, false, granted, 1},
		{
			name:           "revoke a stored and a missing relationship",
			write:          func() error { return writer.Revoke(ctx, link, siteBanRelationship(userID)) },
			wantRows:       afterRevoke,
			wantStatements: 1,
		},
		{
			name:     "grant a batch with an invalid relationship last",
			write:    func() error { return writer.Grant(ctx, owner, invalid) },
			wantErr:  true,
			wantRows: afterRevoke,
		},
		{
			name:     "revoke a batch with an invalid relationship last",
			write:    func() error { return writer.Revoke(ctx, admin, invalid) },
			wantErr:  true,
			wantRows: afterRevoke,
		},
		{"grant nothing", func() error { return writer.Grant(ctx) }, false, afterRevoke, 0},
		{"revoke nothing", func() error { return writer.Revoke(ctx) }, false, afterRevoke, 0},
	}
	for _, step := range steps {
		statements.take()

		err := step.write()

		sent := statements.take()
		if (err != nil) != step.wantErr {
			t.Errorf("%s: error = %v, want error %t", step.name, err, step.wantErr)
		}
		if stored := storedRelationships(t, db, production); !slices.Equal(stored, step.wantRows) {
			t.Errorf("%s: stored = %q, want %q", step.name, stored, step.wantRows)
		}
		if len(sent) != step.wantStatements {
			t.Errorf("%s: sent %d statements, want %d: %q", step.name, len(sent), step.wantStatements, sent)
		}
	}

	t.Run("missing tenant", func(t *testing.T) {
		errs := []error{
			writer.Grant(t.Context(), admin),
			writer.Grant(t.Context()),
			writer.Revoke(t.Context(), admin),
			writer.Revoke(t.Context()),
		}

		for _, err := range errs {
			if errx.KindOf(err) != errx.Unavailable {
				t.Errorf("write without tenant error = %v, want unavailable", err)
			}
		}
	})
}

// TestWriteProperties checks, for random graphs, that Grant is idempotent, that revoking newly granted
// relationships restores the rows, and that every write is one statement.
func TestWriteProperties(t *testing.T) {
	t.Parallel()

	db := newTestDatabase(t)
	pool, statements := newLoggedPool(t, db)
	writer := permissions.NewRelationships(NewRepository(pool), testModel)
	graphs := newGraphPool()

	for _, seed := range propertySeeds {
		key := fmt.Sprintf("e2e/write-properties-%d", seed)
		ctx := registerTestTenant(t, db, key)
		r := rand.New(rand.NewPCG(seed, 0))
		base := graphs.randomRelationships(r, r.IntN(31))
		again := slices.Clone(base)
		r.Shuffle(len(again), func(i, j int) { again[i], again[j] = again[j], again[i] })
		var added []permissions.Relationship
		for _, relationship := range graphs.randomRelationships(r, 1+r.IntN(15)) {
			if !slices.Contains(base, relationship) {
				added = append(added, relationship)
			}
		}
		check := func(step string, err error, batch []permissions.Relationship, want []string) {
			t.Helper()

			sent := statements.take()
			if err != nil {
				propertyFailure(t, seed, base, "%s: %v", step, err)
			}
			if stored := storedRelationships(t, db, key); !slices.Equal(stored, want) {
				propertyFailure(t, seed, base, "%s: stored rows %s", step, rowDifference(stored, want))
			}
			if wantStatements := min(len(batch), 1); len(sent) != wantStatements {
				propertyFailure(t, seed, base, "%s: sent %d statements, want %d", step, len(sent), wantStatements)
			}
		}
		statements.take()

		check("grant", writer.Grant(ctx, base...), base, rows(base))
		check("grant again", writer.Grant(ctx, again...), again, rows(base))
		check("grant more", writer.Grant(ctx, added...), added, rows(append(slices.Clone(base), added...)))
		check("revoke them", writer.Revoke(ctx, added...), added, rows(base))
		check("revoke them again", writer.Revoke(ctx, added...), added, rows(base))
	}
}
