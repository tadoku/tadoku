package repositories

import (
	"time"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/rdb"
	"github.com/tadoku/api/usecases"
)

// NewRankingRepository instantiates a new ranking repository
func NewRankingRepository(sqlHandler rdb.SQLHandler) usecases.RankingRepository {
	return &rankingRepository{sqlHandler: sqlHandler}
}

type rankingRepository struct {
	sqlHandler rdb.SQLHandler
}

func (r *rankingRepository) Store(ranking domain.Ranking) error {
	if ranking.ID == 0 {
		return r.create(ranking)
	}

	return r.update(ranking)
}

func (r *rankingRepository) create(ranking domain.Ranking) error {
	query := `
		insert into rankings
		(contest_id, user_id, language_code, amount, created_at, updated_at)
		values (:contest_id, :user_id, :language_code, :amount, now() at time zone 'utc', now() at time zone 'utc')
	`

	_, err := r.sqlHandler.NamedExecute(query, ranking)
	return domain.WrapError(err)
}

func (r *rankingRepository) update(ranking domain.Ranking) error {
	query := `
		update rankings
		set amount = :amount, updated_at = now() at time zone 'utc'
		where id = :id
	`

	_, err := r.sqlHandler.NamedExecute(query, ranking)
	return domain.WrapError(err)
}

func (r *rankingRepository) UpdateAmounts(rankings domain.Rankings) error {
	tx, err := r.sqlHandler.Begin()
	if err != nil {
		return domain.WrapError(err)
	}

	query := `
		update rankings
		set
			amount = :amount,
			updated_at = now() at time zone 'utc'
		where id = :id
	`

	for _, ranking := range rankings {
		_, err := tx.NamedExecute(query, ranking)

		if err != nil {
			_ = tx.Rollback()
			return domain.WrapError(err)
		}
	}

	return tx.Commit()
}

func (r *rankingRepository) RankingsForContest(
	contestID uint64,
	languageCode domain.LanguageCode,
) (domain.Rankings, error) {
	var rankings []domain.Ranking

	query := `
		select rankings.id as id, contest_id, user_id, language_code, amount, created_at, updated_at, users.display_name as user_display_name
		from rankings
		inner join users on users.id = rankings.user_id
		where contest_id = $1 and language_code = $2
		order by amount desc, id asc
	`

	err := r.sqlHandler.Select(&rankings, query, contestID, languageCode)
	if err != nil {
		return nil, err
	}

	return rankings, nil
}

func (r *rankingRepository) GlobalRankings(languageCode domain.LanguageCode) (domain.Rankings, error) {
	var rankings []domain.Ranking

	query := `
		select user_id, language_code, sum(amount) as amount, users.display_name as user_display_name
		from rankings
		inner join users on users.id = rankings.user_id
		where language_code = $1
		group by user_id, user_display_name, language_code
		order by amount desc
	`

	err := r.sqlHandler.Select(&rankings, query, languageCode)
	if err != nil {
		return nil, err
	}

	return rankings, nil
}

func (r *rankingRepository) FindAll(contestID uint64, userID uint64) (domain.Rankings, error) {
	var rankings []domain.Ranking

	query := `
		select r.id, contest_id, user_id, u.display_name as user_display_name, language_code, amount, created_at, updated_at
		from rankings as r
		inner join users as u on u.id = r.user_id
		where contest_id = $1 and user_id = $2
		order by id asc
	`

	err := r.sqlHandler.Select(&rankings, query, contestID, userID)
	if err != nil {
		return nil, err
	}

	return rankings, nil
}

func (r *rankingRepository) GetAllLanguagesForContestAndUser(contestID uint64, userID uint64) (domain.LanguageCodes, error) {
	var codes []domain.LanguageCode

	query := `
		select language_code
		from rankings
		where
			contest_id = $1
			and user_id = $2
			and language_code != $3
	`

	err := r.sqlHandler.Select(&codes, query, contestID, userID, domain.Global)
	if err != nil {
		return nil, err
	}

	return codes, nil
}

func (r *rankingRepository) CurrentRegistration(userID uint64) (domain.RankingRegistration, error) {
	var rows []struct {
		ID           uint64
		Start        time.Time
		End          time.Time
		LanguageCode domain.LanguageCode `db:"language_code"`
	}

	query := `
		select
			contests.id as id,
			contests."end" as "end",
			contests.start as start,
			rankings.language_code as language_code
		from rankings
		inner join contests on contests.id = rankings.contest_id and contests.open = true
		where rankings.user_id = $1 and rankings.language_code != 'GLO'
	`

	err := r.sqlHandler.Select(&rows, query, userID)
	if err != nil {
		return domain.RankingRegistration{}, domain.WrapError(err)
	}

	result := &domain.RankingRegistration{}

	for i, row := range rows {
		if i == 0 {
			result.ContestID = row.ID
			result.Start = row.Start
			result.End = row.End
		}
		result.Languages = append(result.Languages, row.LanguageCode)
	}

	return *result, nil
}
