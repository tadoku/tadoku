// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: leaderboard.sql

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const leaderboardForContest = `-- name: LeaderboardForContest :many
with leaderboard as (
  select
    user_id,
    sum(score) as score
  from logs
  inner join contest_logs
    on contest_logs.log_id = logs.id
  where
    contest_logs.contest_id = $3
    and logs.deleted_at is null
    and (logs.language_code = $4 or $4 is null)
    and (logs.log_activity_id = $5::integer or $5 is null)
  group by user_id
), registrations as (
  select
    id,
    user_id,
    user_display_name
  from contest_registrations
  where
    contest_id = $3
    and deleted_at is null
    and ($4 = any(language_codes) or $4 is null)
)
select
  rank() over(order by score desc) as rank,
  registrations.user_id,
  registrations.user_display_name,
  coalesce(leaderboard.score, 0)::real as score,
  (select count(registrations.user_id) from registrations) as total_size
from registrations
left join leaderboard using(user_id)
order by
  score desc,
  registrations.user_id asc
limit $2
offset $1
`

type LeaderboardForContestParams struct {
	StartFrom    int32
	PageSize     int32
	ContestID    uuid.UUID
	LanguageCode sql.NullString
	ActivityID   sql.NullInt32
}

type LeaderboardForContestRow struct {
	Rank            int64
	UserID          uuid.UUID
	UserDisplayName string
	Score           float32
	TotalSize       int64
}

func (q *Queries) LeaderboardForContest(ctx context.Context, arg LeaderboardForContestParams) ([]LeaderboardForContestRow, error) {
	rows, err := q.db.QueryContext(ctx, leaderboardForContest,
		arg.StartFrom,
		arg.PageSize,
		arg.ContestID,
		arg.LanguageCode,
		arg.ActivityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LeaderboardForContestRow
	for rows.Next() {
		var i LeaderboardForContestRow
		if err := rows.Scan(
			&i.Rank,
			&i.UserID,
			&i.UserDisplayName,
			&i.Score,
			&i.TotalSize,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
