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

const globalLeaderboard = `-- name: GlobalLeaderboard :many
with leaderboard as (
  select
    user_id,
    sum(score) as score
  from logs
  inner join contest_logs
    on contest_logs.log_id = logs.id
  where
    eligible_official_leaderboard = true
    and logs.deleted_at is null
    and (logs.language_code = $3 or $3 is null)
    and (logs.log_activity_id = $4::integer or $4 is null)
  group by user_id
), ranked_leaderboard as (
  select
    user_id,
    score,
    rank() over(order by score desc) as "rank"
  from leaderboard
  where score > 0
), registrations as (
  select
    contest_registrations.user_id,
    max(contest_registrations.user_display_name)::varchar as user_display_name
  from contest_registrations
  where
    contest_registrations.deleted_at is null
    and ($3 = any(language_codes) or $3 is null)
  group by contest_registrations.user_id
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    registrations.user_id::uuid as user_id,
    registrations.user_display_name::varchar as user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score,
    (select count(registrations.user_id) from registrations) as total_size
  from ranked_leaderboard
  left join registrations using(user_id)
  where
    registrations.user is not null
    and registrations.user_display_name is not null
  order by
    score desc,
    registrations.user_display_name asc
)
select
  rank, user_id, user_display_name, score, total_size,
  coalesce((
    "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
    or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
  ), false)::boolean as is_tie
from enriched_leaderboard
limit $2
offset $1
`

type GlobalLeaderboardParams struct {
	StartFrom    int32
	PageSize     int32
	LanguageCode sql.NullString
	ActivityID   sql.NullInt32
}

type GlobalLeaderboardRow struct {
	Rank            int64
	UserID          uuid.UUID
	UserDisplayName string
	Score           float32
	TotalSize       int64
	IsTie           bool
}

func (q *Queries) GlobalLeaderboard(ctx context.Context, arg GlobalLeaderboardParams) ([]GlobalLeaderboardRow, error) {
	rows, err := q.db.QueryContext(ctx, globalLeaderboard,
		arg.StartFrom,
		arg.PageSize,
		arg.LanguageCode,
		arg.ActivityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GlobalLeaderboardRow
	for rows.Next() {
		var i GlobalLeaderboardRow
		if err := rows.Scan(
			&i.Rank,
			&i.UserID,
			&i.UserDisplayName,
			&i.Score,
			&i.TotalSize,
			&i.IsTie,
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
), ranked_leaderboard as (
  select
    user_id,
    score,
    rank() over(order by score desc) as "rank"
  from leaderboard
), registrations as (
  select
    id,
    user_id,
    user_display_name,
    created_at
  from contest_registrations
  where
    contest_id = $3
    and deleted_at is null
    and ($4 = any(language_codes) or $4 is null)
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    registrations.user_id,
    registrations.user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score,
    (select count(registrations.user_id) from registrations) as total_size
  from registrations
  left join ranked_leaderboard using(user_id)
  order by
    score desc,
    registrations.created_at asc
)
select
  rank, user_id, user_display_name, score, total_size,
  coalesce((
    "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
    or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
  ), false)::boolean as is_tie
from enriched_leaderboard
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
	IsTie           bool
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
			&i.IsTie,
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

const yearlyLeaderboard = `-- name: YearlyLeaderboard :many
with leaderboard as (
  select
    user_id,
    sum(score) as score
  from logs
  inner join contest_logs
    on contest_logs.log_id = logs.id
  where
    logs.year = $3
    and eligible_official_leaderboard = true
    and logs.deleted_at is null
    and (logs.language_code = $4 or $4 is null)
    and (logs.log_activity_id = $5::integer or $5 is null)
  group by user_id
), ranked_leaderboard as (
  select
    user_id,
    score,
    rank() over(order by score desc) as "rank"
  from leaderboard
  where score > 0
), registrations as (
  select
    contest_registrations.id,
    contest_registrations.user_id,
    contest_registrations.user_display_name,
    contest_registrations.created_at
  from contest_registrations
  inner join contests
    on contests.id = contest_registrations.contest_id
  where
    extract(year from contests.contest_start) = $3::integer
    and contest_registrations.deleted_at is null
    and ($4 = any(language_codes) or $4 is null)
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    registrations.user_id::uuid as user_id,
    registrations.user_display_name::varchar as user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score,
    (select count(registrations.user_id) from registrations) as total_size
  from ranked_leaderboard
  left join registrations using(user_id)
  where
    registrations.user is not null
    and registrations.user_display_name is not null
  order by
    score desc,
    registrations.created_at asc
)
select
  rank, user_id, user_display_name, score, total_size,
  coalesce((
    "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
    or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
  ), false)::boolean as is_tie
from enriched_leaderboard
limit $2
offset $1
`

type YearlyLeaderboardParams struct {
	StartFrom    int32
	PageSize     int32
	Year         int16
	LanguageCode sql.NullString
	ActivityID   sql.NullInt32
}

type YearlyLeaderboardRow struct {
	Rank            int64
	UserID          uuid.UUID
	UserDisplayName string
	Score           float32
	TotalSize       int64
	IsTie           bool
}

func (q *Queries) YearlyLeaderboard(ctx context.Context, arg YearlyLeaderboardParams) ([]YearlyLeaderboardRow, error) {
	rows, err := q.db.QueryContext(ctx, yearlyLeaderboard,
		arg.StartFrom,
		arg.PageSize,
		arg.Year,
		arg.LanguageCode,
		arg.ActivityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []YearlyLeaderboardRow
	for rows.Next() {
		var i YearlyLeaderboardRow
		if err := rows.Scan(
			&i.Rank,
			&i.UserID,
			&i.UserDisplayName,
			&i.Score,
			&i.TotalSize,
			&i.IsTie,
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
