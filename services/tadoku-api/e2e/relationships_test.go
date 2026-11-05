package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

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

// Test safety: every reset seeds this production baseline for subjects outside the cast and every case's seeds, so
// parallel scenarios can prove a real production grant or ban stays in production without writing tadoku/prod. They
// only read it, through requireProductionBaseline.
var (
	productionAdmin = fixtureRelationship{
		Tenant:   "tadoku/prod",
		Object:   "app:root",
		Relation: "admins",
		Subject:  "user:aa000000-0000-4000-8000-000000000001",
	}
	productionBan = fixtureRelationship{
		Tenant:   "tadoku/prod",
		Object:   "app:root",
		Relation: "banned",
		Subject:  "user:bb000000-0000-4000-8000-000000000002",
	}
)

// resetRelationships replaces every tenant's stored relationships with the production baseline and the rows in
// seedFiles, in one transaction on the owner pool. Run it after every PostgreSQL reset, after the seeds that register
// the rows' tenants. Missing files are skipped.
func resetRelationships(ctx context.Context, pool *pgxpool.Pool, seedFiles ...string) error {
	batch := &pgx.Batch{}
	batch.Queue("delete from permission_relationships")
	for _, relationship := range []fixtureRelationship{productionAdmin, productionBan} {
		relationship.queue(batch, relationship.Tenant)
	}

	for _, seedFile := range seedFiles {
		relationships, err := readRelationships(seedFile)
		if err != nil {
			return err
		}
		for index, relationship := range relationships {
			if relationship.Tenant == "" {
				return fmt.Errorf("relationship seed %s[%d] needs a tenant", seedFile, index)
			}
			relationship.queue(batch, relationship.Tenant)
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

// readRelationships returns no rows for a missing file. It leaves the tenant to the caller.
func readRelationships(seedFile string) ([]fixtureRelationship, error) {
	contents, err := os.ReadFile(seedFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read relationship seed %s: %w", seedFile, err)
	}

	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var relationships []fixtureRelationship
	if err := decoder.Decode(&relationships); err != nil {
		return nil, fmt.Errorf("decode relationship seed %s: %w", seedFile, err)
	}

	for index, relationship := range relationships {
		_, _, objectOK := strings.Cut(relationship.Object, ":")
		_, _, subjectOK := strings.Cut(relationship.Subject, ":")
		if !objectOK || !subjectOK || relationship.Relation == "" {
			return nil, fmt.Errorf(
				"relationship seed %s[%d] needs a relation and type:id object and subject",
				seedFile,
				index,
			)
		}
	}
	return relationships, nil
}

func (relationship fixtureRelationship) queue(batch *pgx.Batch, tenant string) {
	objectType, objectID, _ := strings.Cut(relationship.Object, ":")
	subjectType, subjectID, _ := strings.Cut(relationship.Subject, ":")
	batch.Queue(insertRelationshipSQL, tenant, objectType, objectID, relationship.Relation, subjectType, subjectID)
}

// requireProductionBaseline fails unless tadoku/prod still holds the baseline relationship, read in a read-only
// transaction, so a case that sends its subject to a test tenant cannot pass without a production grant or ban.
func requireProductionBaseline(t *testing.T, relationship fixtureRelationship) {
	t.Helper()
	objectType, objectID, _ := strings.Cut(relationship.Object, ":")
	subjectType, subjectID, _ := strings.Cut(relationship.Subject, ":")

	var held bool
	err := pgx.BeginTxFunc(t.Context(), api.db.Pool, pgx.TxOptions{AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		return tx.QueryRow(t.Context(), `select exists (
			select from permission_relationships
			where tenant = $1
				and object_type = $2
				and object_id = $3
				and relation = $4
				and subject_type = $5
				and subject_id = $6
		)`,
			relationship.Tenant, objectType, objectID, relationship.Relation, subjectType, subjectID,
		).Scan(&held)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !held {
		t.Fatalf("production baseline %+v is missing; a serial test reset PostgreSQL without resetRelationships", relationship)
	}
}
