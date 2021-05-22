package repositories

import (
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/rdb"
	"github.com/tadoku/api/usecases"
)

// NewContestLogRepository instantiates a new contest log repository
func NewContestLogRepository(sqlHandler rdb.SQLHandler) usecases.ContestLogRepository {
	return &contestLogRepository{sqlHandler: sqlHandler}
}

type contestLogRepository struct {
	sqlHandler rdb.SQLHandler
}

func (r *contestLogRepository) Store(contestLog *domain.ContestLog) error {
	if contestLog.ID == 0 {
		return r.create(contestLog)
	}

	return r.update(contestLog)
}

func (r *contestLogRepository) create(contestLog *domain.ContestLog) error {
	query := `
		insert into contest_logs
		(contest_id, user_id, language_code, medium_id, amount, description, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, now() at time zone 'utc', now() at time zone 'utc')
		returning id
	`

	row := r.sqlHandler.QueryRow(
		query,
		contestLog.ContestID,
		contestLog.UserID,
		contestLog.Language,
		contestLog.MediumID,
		contestLog.Amount,
		contestLog.Description,
	)
	err := row.Scan(&contestLog.ID)
	if err != nil {
		return domain.WrapError(err)
	}

	return nil
}

func (r *contestLogRepository) update(contestLog *domain.ContestLog) error {
	query := `
		update contest_logs
		set amount = :amount, medium_id = :medium_id, language_code = :language_code, description = :description, updated_at = now() at time zone 'utc'
		where
			id = :id and
			user_id = :user_id and
			deleted_at is null
	`

	_, err := r.sqlHandler.NamedExecute(query, contestLog)
	return domain.WrapError(err)
}

func (r *contestLogRepository) FindByID(id uint64) (domain.ContestLog, error) {
	l := domain.ContestLog{}

	query := `
		select id, contest_id, user_id, language_code, medium_id, amount, description, created_at, updated_at
		from contest_logs
		where
			id = $1 and
			deleted_at is null
	`
	err := r.sqlHandler.QueryRow(query, id).StructScan(&l)
	if err != nil {
		return l, domain.WrapError(err)
	}
	if l.ID == 0 {
		return l, domain.ErrNotFound
	}

	return l, nil
}

func (r *contestLogRepository) FindAll(contestID uint64, userID uint64) (domain.ContestLogs, error) {
	var logs []domain.ContestLog

	query := `
		select
			l.id,
			l.contest_id,
			l.user_id,
			l.language_code,
			l.medium_id,
			l.amount,
			l.description,
			l.created_at,
			l.updated_at,
			coalesce(u.display_name, '') as user_display_name
		from contest_logs as l
		left join users as u on u.id = l.user_id
		where
			l.contest_id = $1 and
			l.user_id = $2 and
			l.deleted_at is null
	`

	err := r.sqlHandler.Select(&logs, query, contestID, userID)
	if err != nil {
		return nil, domain.WrapError(err)
	}

	return logs, nil
}

func (r *contestLogRepository) Delete(id uint64) error {
	query := `
		update contest_logs
		set deleted_at = now() at time zone 'utc'
		where
			id = $1 and
			deleted_at is null
	`

	result, err := r.sqlHandler.Execute(query, id)
	if err != nil {
		return domain.WrapError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.WrapError(err)
	}
	if rows != 1 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *contestLogRepository) FindRecent(contestID, limit uint64) (domain.ContestLogs, error) {
	var logs []domain.ContestLog

	query := `
		select
			l.id,
			l.contest_id,
			l.user_id,
			l.language_code,
			l.medium_id,
			l.amount,
			l.description,
			l.created_at,
			l.updated_at,
			coalesce(u.display_name, '') as user_display_name
		from contest_logs as l
		left join users as u on u.id = l.user_id
		where
			l.contest_id = $1 and
			l.deleted_at is null
		order by l.created_at desc, l.id desc
		limit $2
	`

	err := r.sqlHandler.Select(&logs, query, contestID, limit)
	if err != nil {
		return nil, domain.WrapError(err)
	}

	return logs, nil
}
