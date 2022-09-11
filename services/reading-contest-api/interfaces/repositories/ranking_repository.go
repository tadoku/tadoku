package repositories

import (
	"time"

	"github.com/lib/pq"

	"github.com/tadoku/tadoku/services/reading-contest-api/domain"
	"github.com/tadoku/tadoku/services/reading-contest-api/interfaces/rdb"
	"github.com/tadoku/tadoku/services/reading-contest-api/usecases"
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
		(contest_id, user_id, language_code, amount, user_display_name, created_at, updated_at)
		values (:contest_id, :user_id, :language_code, :amount, :user_display_name, now() at time zone 'utc', now() at time zone 'utc')
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

	// TODO: remove filter on language code once they're reworked
	// TODO: rename rankings
	query := `
		with leaderboard as (
			select
				user_id,
				sum(weighted_score) as amount
			from contest_logs
			where contest_id = $1
			group by user_id
		), registrations as (
			select
				min(id) as id,
				user_id,
				array_agg(language_code) as language_codes,
				user_display_name
			from rankings
			where contest_id = $1
			group by
				user_id,
				user_display_name
		)

		select
			leaderboard.user_id,
			leaderboard.amount,
			registrations.user_display_name,
			'GLO' as language_code,
			$1 as contest_id
			-- , registrations.language_codes
		from leaderboard
		inner join registrations using(user_id)
		order by
			amount desc,
			registrations.id asc;
	`

	err := r.sqlHandler.Select(&rankings, query, contestID)
	if err != nil {
		return nil, err
	}

	return rankings, nil
}

func (r *rankingRepository) GlobalRankings(languageCode domain.LanguageCode) (domain.Rankings, error) {
	var rankings []domain.Ranking

	query := `
		select user_id, language_code, sum(amount) as amount, user_display_name
		from rankings
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
		select r.id, contest_id, user_id, user_display_name, language_code, amount, created_at, updated_at
		from rankings as r
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

func (r *rankingRepository) CurrentRegistration(userID uint64, now time.Time) (domain.RankingRegistration, error) {
	type Row struct {
		ContestID uint64 `db:"contest_id"`
		Start     time.Time
		End       time.Time
		Languages pq.StringArray `db:"language_codes"`
	}
	row := Row{}
	result := &domain.RankingRegistration{}

	query := `
		select
			contests.id as contest_id,
			contests."end" as "end",
			contests.start as start,
			array_agg(distinct rankings.language_code) as language_codes
		from contests
		left join rankings on contests.id = rankings.contest_id
		where
			rankings.user_id = $1 and
			rankings.language_code != 'GLO' and
			$2::date <= contests."end"
		group by contests.id
		order by contests.id desc
		limit 1
	`

	err := r.sqlHandler.Get(&row, query, userID, now)
	if err != nil {
		return domain.RankingRegistration{}, domain.WrapError(err)
	}

	result.ContestID = row.ContestID
	result.Start = row.Start
	result.End = row.End

	for _, code := range row.Languages {
		result.Languages = append(result.Languages, domain.LanguageCode(code))
	}

	return *result, nil
}
