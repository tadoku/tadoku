package repositories

import (
	"github.com/srvc/fail"
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
		(contest_id, user_id, language_code, medium_id, amount, created_at, updated_at)
		values (:contest_id, :user_id, :language_code, :medium_id, :amount, now() at time zone 'utc', now() at time zone 'utc')
	`

	_, err := r.sqlHandler.NamedExecute(query, contestLog)
	return fail.Wrap(err)
}

func (r *contestLogRepository) update(contestLog *domain.ContestLog) error {
	query := `
		update contest_logs
		set amount = :amount, medium_id = :medium_id, language_code = :language_code, updated_at = now() at time zone 'utc'
		where id = :id
	`

	_, err := r.sqlHandler.NamedExecute(query, contestLog)
	return fail.Wrap(err)
}

func (r *contestLogRepository) FindAll(contestID uint64, userID uint64) (domain.ContestLogs, error) {
	var logs []domain.ContestLog

	query := `
		select
			id, contest_id, user_id, language_code, medium_id, amount, created_at, updated_at
		from contest_logs
		where contest_id = $1 and user_id = $2
	`

	err := r.sqlHandler.Select(&logs, query, contestID, userID)
	if err != nil {
		return nil, fail.Wrap(err)
	}

	return logs, nil
}
