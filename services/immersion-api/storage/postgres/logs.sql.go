// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: logs.sql

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const createContestLogRelation = `-- name: CreateContestLogRelation :exec
insert into contest_logs (
  contest_id,
  log_id
) values (
  (select contest_id from contest_registrations where id = $1),
  $2
)
`

type CreateContestLogRelationParams struct {
	RegistrationID uuid.UUID
	LogID          uuid.UUID
}

func (q *Queries) CreateContestLogRelation(ctx context.Context, arg CreateContestLogRelationParams) error {
	_, err := q.db.ExecContext(ctx, createContestLogRelation, arg.RegistrationID, arg.LogID)
	return err
}

const createLog = `-- name: CreateLog :one
insert into logs (
  id,
  user_id,
  language_code,
  log_activity_id,
  unit_id,
  tags,
  amount,
  modifier,
  eligible_official_leaderboard,
  "description"
) values (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10
) returning id
`

type CreateLogParams struct {
	ID                          uuid.UUID
	UserID                      uuid.UUID
	LanguageCode                string
	LogActivityID               int16
	UnitID                      uuid.UUID
	Tags                        []string
	Amount                      float32
	Modifier                    float32
	EligibleOfficialLeaderboard bool
	Description                 sql.NullString
}

func (q *Queries) CreateLog(ctx context.Context, arg CreateLogParams) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, createLog,
		arg.ID,
		arg.UserID,
		arg.LanguageCode,
		arg.LogActivityID,
		arg.UnitID,
		pq.Array(arg.Tags),
		arg.Amount,
		arg.Modifier,
		arg.EligibleOfficialLeaderboard,
		arg.Description,
	)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}