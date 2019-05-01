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

func (r *contestLogRepository) Store(contestLog domain.ContestLog) error {
	if contestLog.ID == 0 {
		return r.create(contestLog)
	}

	return fail.Errorf("not yet implemented")
}

func (r *contestLogRepository) create(contestLog domain.ContestLog) error {
	query := `
		insert into contest_logs
		(contest_id, user_id, language_code, medium_id, amount, created_at, updated_at)
		values (:contest_id, :user_id, :language_code, :medium_id, :amount, now(), now())
	`

	_, err := r.sqlHandler.NamedExecute(query, contestLog)
	return fail.Wrap(err)
}

func (r *contestLogRepository) FindAll(contestID uint64, userID uint64) (domain.ContestLogs, error) {
	return nil, nil
}
