package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// fixtureRelationship is one permission_relationships row in a relationships.json seed file. Object and subject
// are type:id, for example app:root and user:<uuid>.
type fixtureRelationship struct {
	Tenant   string `json:"tenant"`
	Object   string `json:"object"`
	Relation string `json:"relation"`
	Subject  string `json:"subject"`
}

const insertRelationshipSQL = `insert into permission_relationships
	(tenant, object_type, object_id, relation, subject_type, subject_id)
	values ($1, $2, $3, $4, $5, $6)`

// resetRelationships replaces every tenant's stored relationships with the rows in seedFiles, in one transaction on
// the owner pool. Run it after the PostgreSQL seeds that register the rows' tenants. Missing files are skipped.
func resetRelationships(ctx context.Context, pool *pgxpool.Pool, seedFiles ...string) error {
	batch := &pgx.Batch{}
	batch.Queue("delete from permission_relationships")

	for _, seedFile := range seedFiles {
		contents, err := os.ReadFile(seedFile)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read relationship seed %s: %w", seedFile, err)
		}

		decoder := json.NewDecoder(bytes.NewReader(contents))
		decoder.DisallowUnknownFields()
		var relationships []fixtureRelationship
		if err := decoder.Decode(&relationships); err != nil {
			return fmt.Errorf("decode relationship seed %s: %w", seedFile, err)
		}

		for index, relationship := range relationships {
			objectType, objectID, objectOK := strings.Cut(relationship.Object, ":")
			subjectType, subjectID, subjectOK := strings.Cut(relationship.Subject, ":")
			if !objectOK || !subjectOK || relationship.Tenant == "" || relationship.Relation == "" {
				return fmt.Errorf(
					"relationship seed %s[%d] needs a tenant, a relation and type:id object and subject",
					seedFile,
					index,
				)
			}
			batch.Queue(
				insertRelationshipSQL,
				relationship.Tenant,
				objectType,
				objectID,
				relationship.Relation,
				subjectType,
				subjectID,
			)
		}
	}

	err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		return tx.SendBatch(ctx, batch).Close()
	})
	if err != nil {
		return fmt.Errorf("seed relationships: %w", err)
	}
	return nil
}
