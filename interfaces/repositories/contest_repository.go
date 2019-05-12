package repositories

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/rdb"
	"github.com/tadoku/api/usecases"
)

// NewContestRepository instantiates a new contest repository
func NewContestRepository(sqlHandler rdb.SQLHandler) usecases.ContestRepository {
	return &contestRepository{sqlHandler: sqlHandler}
}

type contestRepository struct {
	sqlHandler rdb.SQLHandler
}

func (r *contestRepository) Store(contest *domain.Contest) error {
	if contest.ID == 0 {
		return r.create(contest)
	}

	return r.update(contest)
}

func (r *contestRepository) create(contest *domain.Contest) error {
	query := `
		insert into contests
		(start, "end", open)
		values ($1, $2, $3)
		returning id
	`

	row := r.sqlHandler.QueryRow(query, contest.Start, contest.End, contest.Open)
	err := row.Scan(&contest.ID)
	if err != nil {
		return fail.Wrap(err)
	}

	return nil
}

func (r *contestRepository) update(contest *domain.Contest) error {
	query := `
		update contests
		set start = :start, "end" = :end, open = :open
		where id = :id
	`

	_, err := r.sqlHandler.NamedExecute(query, contest)
	return fail.Wrap(err)
}

func (r *contestRepository) GetOpenContests() ([]uint64, error) {
	query := `
		select id
		from contests
		where open = true
	`

	var ids []uint64
	err := r.sqlHandler.Select(&ids, query)
	if err != nil {
		return nil, fail.Wrap(err)
	}

	return ids, nil
}
