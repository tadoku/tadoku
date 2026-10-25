package authz

import (
	"context"
	"maps"
	"slices"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"pgregory.net/rapid"
)

// TestRelationshipStoreProperties runs random sequences of Grant, Revoke, RevokeObject, RevokeUser and Read against
// Postgres. After every step the stored rows equal the expected rows. Each call returns what the expected rows
// predict, rejects invalid input, and sends one statement, or none when it has nothing to do or rejects its input.
func TestRelationshipStoreProperties(t *testing.T) {
	t.Parallel()

	db := newTestDatabase(t)
	pool, statements := newLoggedPool(t, db)
	store := permissions.NewRelationships(NewRepository(pool), testModel)
	production := tenant.Production().String()
	graphs := newGraphPool()
	writes, reads := graphs.writes(), graphs.reads()
	invalid, invalidObjects := graphs.invalidRelationships(), []permissions.Object{{}, contests.Object("")}
	objects := rapid.SampledFrom(append(graphs.objects(), invalidObjects...))
	users := rapid.IntRange(0, len(graphs.users)-1)

	rapid.Check(t, func(t *rapid.T) {
		ctx := tenant.WithKey(t.Context(), tenant.Production())
		clearRelationships(t, db, production)
		expected := map[string]permissions.Relationship{}
		call := func(t *rapid.T, rejected bool, wantStatements int, run func() error) {
			t.Helper()
			statements.take()

			err := run()

			sent := statements.take()
			if (err != nil) != rejected {
				t.Fatalf("error = %v, want rejected %t", err, rejected)
			}
			if rejected {
				wantStatements = 0
			}
			if len(sent) != wantStatements {
				t.Fatalf("sent %d statements, want %d: %q", len(sent), wantStatements, sent)
			}
		}
		write := func(
			apply func(context.Context, ...permissions.Relationship) error,
			update func(permissions.Relationship),
		) func(*rapid.T) {
			return func(t *rapid.T) {
				relationships := writes.Draw(t, "relationships")
				rejected := slices.ContainsFunc(relationships, func(r permissions.Relationship) bool {
					return slices.Contains(invalid, r)
				})

				call(t, rejected, min(len(relationships), 1), func() error {
					return apply(ctx, relationships...)
				})

				if !rejected {
					for _, relationship := range relationships {
						update(relationship)
					}
				}
			}
		}

		t.Repeat(map[string]func(*rapid.T){
			"Grant":  write(store.Grant, func(r permissions.Relationship) { expected[row(r)] = r }),
			"Revoke": write(store.Revoke, func(r permissions.Relationship) { delete(expected, row(r)) }),
			"RevokeObject": func(t *rapid.T) {
				object := objects.Draw(t, "object")
				rejected := slices.Contains(invalidObjects, object)

				call(t, rejected, 1, func() error { return store.RevokeObject(ctx, object) })

				if !rejected {
					maps.DeleteFunc(expected, func(_ string, relationship permissions.Relationship) bool {
						return names(relationship, object)
					})
				}
			},
			"RevokeUser": func(t *rapid.T) {
				user := graphs.users[users.Draw(t, "user")]

				call(t, false, 1, func() error { return store.RevokeUser(ctx, user) })

				maps.DeleteFunc(expected, func(_ string, relationship permissions.Relationship) bool {
					return storesUser(user)(relationship)
				})
			},
			"Read": func(t *rapid.T) {
				read := reads.Draw(t, "filter")
				var got []permissions.Relationship

				call(t, false, 1, func() (err error) {
					got, err = store.Read(ctx, read.filter)
					return err
				})

				var want []string
				for row, relationship := range expected {
					if read.selects(relationship) {
						want = append(want, row)
					}
				}
				slices.Sort(want)
				gotRows := make([]string, len(got))
				for i, relationship := range got {
					gotRows[i] = row(relationship)
				}
				slices.Sort(gotRows)
				if !slices.Equal(gotRows, want) {
					t.Fatalf("read rows %s", rowDifference(gotRows, want))
				}
				for _, relationship := range got {
					if expected[row(relationship)] != relationship {
						t.Fatalf("read %s, which is not equal to the relationship granted", row(relationship))
					}
				}
			},
			"": func(t *rapid.T) {
				want := slices.Sorted(maps.Keys(expected))
				if stored := storedRelationships(t, db, production); !slices.Equal(stored, want) {
					t.Fatalf("stored rows %s", rowDifference(stored, want))
				}
			},
		})
	})
}

// names reports whether the relationship is on the object or names it as a linked object or in a subject set.
func names(relationship permissions.Relationship, object permissions.Object) bool {
	target, _ := relationship.Target()
	set, _, _ := relationship.Set()
	return relationship.Object() == object || target == object || set == object
}
