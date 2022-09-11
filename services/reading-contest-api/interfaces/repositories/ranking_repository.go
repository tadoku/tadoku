package repositories

import (
	"errors"
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
		(contest_id, user_id, language_code, user_display_name, created_at, updated_at)
		values (:contest_id, :user_id, :language_code, :user_display_name, now() at time zone 'utc', now() at time zone 'utc')
	`

	_, err := r.sqlHandler.NamedExecute(query, ranking)
	return domain.WrapError(err)
}

func (r *rankingRepository) update(ranking domain.Ranking) error {
	return errors.New("rankings cannot be updated at this time")
}

func (r *rankingRepository) RankingsForContest(
	contestID uint64,
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
			where
				contest_id = $1 and
				deleted_at is null
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
			coalesce(leaderboard.amount, 0) as amount,
			registrations.user_id,
			registrations.user_display_name,
			'GLO' as language_code,
			$1 as contest_id
			-- , registrations.language_codes
		from registrations
		left join leaderboard using(user_id)
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

func (r *rankingRepository) FindAll(contestID uint64, userID uint64) (domain.Rankings, error) {
	var rankings []domain.Ranking

	query := `
		with scores as (
			select
				language_code,
				sum(weighted_score) as amount
			from contest_logs
			where contest_id = $1 and user_id = $2
			group by language_code
		)

		select
			r.id,
			contest_id,
			user_id,
			user_display_name,
			r.language_code,
			created_at,
			updated_at,
			coalesce(s.amount, 0) as amount
		from rankings as r
		left join scores as s on (s.language_code = r.language_code)
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
