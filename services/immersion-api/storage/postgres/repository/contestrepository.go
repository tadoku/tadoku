package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

// COMMANDS

func (r *Repository) CreateContest(ctx context.Context, req *command.ContestCreateRequest) (*command.ContestCreateResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create contest: %w", err)
	}

	qtx := r.q.WithTx(tx)

	id, err := qtx.CreateContest(ctx, postgres.CreateContestParams{
		OwnerUserID:             req.OwnerUserID,
		OwnerUserDisplayName:    req.OwnerUserDisplayName,
		Official:                req.Official,
		Private:                 req.Private,
		ContestStart:            req.ContestStart,
		ContestEnd:              req.ContestEnd,
		RegistrationEnd:         req.RegistrationEnd,
		Title:                   req.Title,
		Description:             postgres.NewNullString(req.Description),
		LanguageCodeAllowList:   req.LanguageCodeAllowList,
		ActivityTypeIDAllowList: req.ActivityTypeIDAllowList,
	})

	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create contest: %w", err)
	}

	contest, err := qtx.FindContestById(ctx, postgres.FindContestByIdParams{
		ID:             id,
		IncludeDeleted: false,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create contest: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not create contest: %w", err)
	}

	return &command.ContestCreateResponse{
		ID:                      contest.ID,
		ContestStart:            contest.ContestStart,
		ContestEnd:              contest.ContestEnd,
		RegistrationEnd:         contest.RegistrationEnd,
		Title:                   contest.Title,
		Description:             postgres.NewStringFromNullString(contest.Description),
		OwnerUserID:             contest.OwnerUserID,
		OwnerUserDisplayName:    contest.OwnerUserDisplayName,
		Official:                contest.Official,
		Private:                 contest.Private,
		LanguageCodeAllowList:   contest.LanguageCodeAllowList,
		ActivityTypeIDAllowList: contest.ActivityTypeIDAllowList,
		CreatedAt:               contest.CreatedAt,
		UpdatedAt:               contest.UpdatedAt,
	}, nil
}

func (r *Repository) UpsertContestRegistration(ctx context.Context, req *command.UpsertContestRegistrationRequest) error {
	_, err := r.q.UpsertContestRegistration(ctx, postgres.UpsertContestRegistrationParams{
		ID:              req.ID,
		ContestID:       req.ContestID,
		UserID:          req.UserID,
		UserDisplayName: req.UserDisplayName,
		LanguageCodes:   req.LanguageCodes,
	})

	if err != nil {
		return fmt.Errorf("could not create or update contest registration: %w", err)
	}

	return nil
}

// QUERIES

func (r *Repository) FetchContestConfigurationOptions(ctx context.Context) (*query.FetchContestConfigurationOptionsResponse, error) {
	langs, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest configuration options: %w", err)
	}

	acts, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest configuration options: %w", err)
	}

	options := query.FetchContestConfigurationOptionsResponse{
		Languages:  make([]query.Language, len(langs)),
		Activities: make([]query.Activity, len(acts)),
	}

	for i, l := range langs {
		options.Languages[i] = query.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	for i, a := range acts {
		options.Activities[i] = query.Activity{
			ID:      a.ID,
			Name:    a.Name,
			Default: a.Default,
		}
	}

	return &options, err
}

func (r *Repository) FindLatestOfficial(ctx context.Context) (*query.ContestView, error) {
	contest, err := r.q.FindLatestOfficialContest(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	activities, err := r.q.ListActivitiesForContest(ctx, contest.ID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	acts := make([]query.Activity, len(activities))
	for i, a := range activities {
		acts[i] = query.Activity{
			ID:   a.ID,
			Name: a.Name,
		}
	}

	langs := []query.Language{}

	if len(contest.LanguageCodeAllowList) > 0 {
		languages, err := r.q.ListLanguagesForContest(ctx, contest.ID)
		if err != nil {
			return nil, fmt.Errorf("could not fetch contest: %w", err)
		}

		langs = make([]query.Language, len(languages))
		for i, a := range languages {
			langs[i] = query.Language{
				Code: a.Code,
				Name: a.Name,
			}
		}
	}

	return &query.ContestView{
		ID:                   contest.ID,
		ContestStart:         contest.ContestStart,
		ContestEnd:           contest.ContestEnd,
		RegistrationEnd:      contest.RegistrationEnd,
		Title:                contest.Title,
		Description:          postgres.NewStringFromNullString(contest.Description),
		OwnerUserID:          contest.OwnerUserID,
		OwnerUserDisplayName: contest.OwnerUserDisplayName,
		Official:             contest.Official,
		Private:              contest.Private,
		AllowedLanguages:     langs,
		AllowedActivities:    acts,
		CreatedAt:            contest.CreatedAt,
		UpdatedAt:            contest.UpdatedAt,
		Deleted:              contest.DeletedAt.Valid,
	}, nil
}

func (r *Repository) FindByID(ctx context.Context, req *query.FindByIDRequest) (*query.ContestView, error) {
	contest, err := r.q.FindContestById(ctx, postgres.FindContestByIdParams{
		ID:             req.ID,
		IncludeDeleted: req.IncludeDeleted,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	activities, err := r.q.ListActivitiesForContest(ctx, contest.ID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	acts := make([]query.Activity, len(activities))
	for i, a := range activities {
		acts[i] = query.Activity{
			ID:   a.ID,
			Name: a.Name,
		}
	}

	langs := []query.Language{}

	if len(contest.LanguageCodeAllowList) > 0 {
		languages, err := r.q.ListLanguagesForContest(ctx, contest.ID)
		if err != nil {
			return nil, fmt.Errorf("could not fetch contest: %w", err)
		}

		langs = make([]query.Language, len(languages))
		for i, a := range languages {
			langs[i] = query.Language{
				Code: a.Code,
				Name: a.Name,
			}
		}
	}

	return &query.ContestView{
		ID:                   contest.ID,
		ContestStart:         contest.ContestStart,
		ContestEnd:           contest.ContestEnd,
		RegistrationEnd:      contest.RegistrationEnd,
		Title:                contest.Title,
		Description:          postgres.NewStringFromNullString(contest.Description),
		OwnerUserID:          contest.OwnerUserID,
		OwnerUserDisplayName: contest.OwnerUserDisplayName,
		Official:             contest.Official,
		Private:              contest.Private,
		AllowedLanguages:     langs,
		AllowedActivities:    acts,
		CreatedAt:            contest.CreatedAt,
		UpdatedAt:            contest.UpdatedAt,
		Deleted:              contest.DeletedAt.Valid,
	}, nil
}

func (r *Repository) ListContests(ctx context.Context, req *query.ContestListRequest) (*query.ContestListResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list contests: %w", err)
	}

	qtx := r.q.WithTx(tx)

	meta, err := qtx.ContestsMetadata(ctx, postgres.ContestsMetadataParams{
		IncludeDeleted: req.IncludeDeleted,
		UserID:         req.UserID,
		Official:       req.OfficialOnly,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not lists contests: %w", err)
	}

	contests, err := qtx.ListContests(ctx, postgres.ListContestsParams{
		StartFrom:      int32(req.Page * req.PageSize),
		PageSize:       int32(req.PageSize),
		IncludeDeleted: req.IncludeDeleted,
		IncludePrivate: req.IncludePrivate,
		UserID:         req.UserID,
		Official:       req.OfficialOnly,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not list contests: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not list contests: %w", err)
	}

	res := make([]query.Contest, len(contests))
	for i, c := range contests {
		res[i] = query.Contest{
			ID:                      c.ID,
			ContestStart:            c.ContestStart,
			ContestEnd:              c.ContestEnd,
			RegistrationEnd:         c.RegistrationEnd,
			Title:                   c.Title,
			Description:             postgres.NewStringFromNullString(c.Description),
			OwnerUserID:             c.OwnerUserID,
			OwnerUserDisplayName:    c.OwnerUserDisplayName,
			Official:                c.Official,
			Private:                 c.Private,
			LanguageCodeAllowList:   c.LanguageCodeAllowList,
			ActivityTypeIDAllowList: c.ActivityTypeIDAllowList,
			CreatedAt:               c.CreatedAt,
			UpdatedAt:               c.UpdatedAt,
			Deleted:                 c.DeletedAt.Valid,
		}
	}

	nextPageToken := ""
	if (req.Page*req.PageSize)+req.PageSize < int(meta.TotalSize) {
		nextPageToken = fmt.Sprint(req.Page + 1)
	}

	return &query.ContestListResponse{
		Contests:      res,
		TotalSize:     int(meta.TotalSize),
		NextPageToken: nextPageToken,
	}, nil
}

func (r *Repository) FindRegistrationForUser(ctx context.Context, req *query.FindRegistrationForUserRequest) (*query.ContestRegistration, error) {
	reg, err := r.q.FindContestRegistrationForUser(ctx, postgres.FindContestRegistrationForUserParams{
		UserID:    req.UserID,
		ContestID: req.ContestID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch contest registration: %w", err)
	}

	languages, err := r.q.GetLanguagesByCode(ctx, reg.LanguageCodes)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest registrations: %w", err)
	}

	registrationLanguages := make([]query.Language, len(reg.LanguageCodes))
	for i, it := range languages {
		registrationLanguages[i] = query.Language{
			Code: it.Code,
			Name: it.Name,
		}
	}

	contest := &query.ContestView{
		ID:                reg.ContestID,
		ContestStart:      reg.ContestStart,
		ContestEnd:        reg.ContestEnd,
		RegistrationEnd:   reg.RegistrationEnd,
		Title:             reg.Title,
		Description:       postgres.NewStringFromNullString(reg.Description),
		Private:           reg.Private,
		Official:          reg.Official,
		AllowedLanguages:  make([]query.Language, 0),
		AllowedActivities: make([]query.Activity, 0),
	}

	return &query.ContestRegistration{
		ID:              reg.ID,
		ContestID:       reg.ContestID,
		UserID:          req.UserID,
		UserDisplayName: reg.UserDisplayName,
		Languages:       registrationLanguages,
		Contest:         contest,
	}, nil
}

func (r *Repository) FetchContestLeaderboard(ctx context.Context, req *query.FetchContestLeaderboardRequest) (*query.Leaderboard, error) {
	_, err := r.q.FindContestById(ctx, postgres.FindContestByIdParams{ID: req.ContestID, IncludeDeleted: false})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch leaderboard for contest: %w", err)
	}

	entries, err := r.q.LeaderboardForContest(ctx, postgres.LeaderboardForContestParams{
		ContestID:    req.ContestID,
		LanguageCode: postgres.NewNullString(req.LanguageCode),
		ActivityID:   postgres.NewNullInt32(req.ActivityID),
		StartFrom:    int32(req.Page * req.PageSize),
		PageSize:     int32(req.PageSize),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.Leaderboard{
				Entries:       []query.LeaderboardEntry{},
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch leaderboard for contest: %w", err)
	}

	res := make([]query.LeaderboardEntry, len(entries))
	for i, e := range entries {
		res[i] = query.LeaderboardEntry{
			Rank:            int(e.Rank),
			UserID:          e.UserID,
			UserDisplayName: e.UserDisplayName,
			Score:           e.Score,
			IsTie:           e.IsTie,
		}
	}

	var totalSize int64
	if len(entries) > 0 {
		totalSize = entries[0].TotalSize
	}
	nextPageToken := ""
	if (req.Page*req.PageSize)+req.PageSize < int(totalSize) {
		nextPageToken = fmt.Sprint(req.Page + 1)
	}

	return &query.Leaderboard{
		Entries:       res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

func (r *Repository) FetchOngoingContestRegistrations(ctx context.Context, req *query.FetchOngoingContestRegistrationsRequest) (*query.ContestRegistrations, error) {
	regs, err := r.q.FindOngoingContestRegistrationForUser(ctx, postgres.FindOngoingContestRegistrationForUserParams{
		UserID: req.UserID,
		Now:    req.Now,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.ContestRegistrations{
				Registrations: []query.ContestRegistration{},
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}
		return nil, fmt.Errorf("could not fetch ongoing contest registrations: %w", err)
	}

	languages, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch ongoing contest registrations: %w", err)
	}

	activities, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch ongoing contest registrations: %w", err)
	}

	langs := map[string]string{}
	acts := map[int32]string{}

	for _, l := range languages {
		langs[l.Code] = l.Name
	}
	for _, a := range activities {
		acts[a.ID] = a.Name
	}

	res := &query.ContestRegistrations{
		Registrations: make([]query.ContestRegistration, len(regs)),
		TotalSize:     len(regs),
		NextPageToken: "",
	}
	for i, r := range regs {
		r := r

		contest := &query.ContestView{
			ID:                r.ContestID,
			ContestStart:      r.ContestStart,
			ContestEnd:        r.ContestEnd,
			RegistrationEnd:   r.RegistrationEnd,
			Title:             r.Title,
			Description:       postgres.NewStringFromNullString(r.Description),
			Private:           r.Private,
			Official:          r.Official,
			AllowedLanguages:  make([]query.Language, 0),
			AllowedActivities: make([]query.Activity, len(r.ActivityTypeIDAllowList)),
		}

		for i, a := range r.ActivityTypeIDAllowList {
			contest.AllowedActivities[i] = query.Activity{
				ID:   a,
				Name: acts[a],
			}
		}

		reg := query.ContestRegistration{
			ID:              r.ID,
			ContestID:       r.ContestID,
			UserID:          r.UserID,
			UserDisplayName: r.UserDisplayName,
			Languages:       make([]query.Language, len(r.LanguageCodes)),
			Contest:         contest,
		}

		for i, code := range r.LanguageCodes {
			reg.Languages[i] = query.Language{
				Code: code,
				Name: langs[code],
			}
		}

		res.Registrations[i] = reg
	}

	return res, nil
}

func (r *Repository) FindScoresForRegistration(ctx context.Context, req *query.ContestProfileRequest) ([]query.Score, error) {
	rows, err := r.q.FetchScoresForContestProfile(ctx, postgres.FetchScoresForContestProfileParams{
		ContestID: req.ContestID,
		UserID:    req.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	scores := make([]query.Score, len(rows))
	for i, row := range rows {
		scores[i] = query.Score{
			LanguageCode: row.LanguageCode,
			Score:        row.Score,
		}
	}

	return scores, nil
}

func (r *Repository) ReadingActivityForContestUser(ctx context.Context, req *query.ContestProfileRequest) ([]query.ReadingActivityRow, error) {
	rows, err := r.q.ReadingActivityPerLanguageForContestProfile(ctx, postgres.ReadingActivityPerLanguageForContestProfileParams{
		ContestID: req.ContestID,
		UserID:    req.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []query.ReadingActivityRow{}, nil
		}
		return nil, fmt.Errorf("could not fetch reading activity: %w", err)
	}

	res := make([]query.ReadingActivityRow, len(rows))
	for i, it := range rows {
		res[i] = query.ReadingActivityRow{
			Date:         it.Date,
			LanguageCode: it.LanguageCode,
			Score:        it.Score,
		}
	}

	return res, nil
}

func (r *Repository) YearlyActivityForUser(ctx context.Context, req *query.YearlyActivityForUserRequest) ([]query.UserActivityScore, error) {
	rows, err := r.q.YearlyActivityForUser(ctx, postgres.YearlyActivityForUserParams{
		UserID: req.UserID,
		Year:   int16(req.Year),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []query.UserActivityScore{}, nil
		}
		return nil, fmt.Errorf("could not fetch activity summary: %w", err)
	}

	res := make([]query.UserActivityScore, len(rows))
	for i, it := range rows {
		res[i] = query.UserActivityScore{
			Date:    it.Date,
			Score:   it.Score,
			Updates: int(it.UpdateCount),
		}
	}

	return res, nil
}

func (r *Repository) YearlyScoresForUser(ctx context.Context, req *query.YearlyScoresForUserRequest) ([]query.Score, error) {
	rows, err := r.q.FetchScoresForProfile(ctx, postgres.FetchScoresForProfileParams{
		UserID: req.UserID,
		Year:   int16(req.Year),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []query.Score{}, nil
		}
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	scores := make([]query.Score, len(rows))
	for i, row := range rows {
		row := row
		scores[i] = query.Score{
			LanguageCode: row.LanguageCode,
			LanguageName: &row.LanguageName,
			Score:        row.Score,
		}
	}

	return scores, nil
}

func (r *Repository) YearlyContestRegistrations(ctx context.Context, req *query.YearlyContestRegistrationsRequest) (*query.ContestRegistrations, error) {
	regs, err := r.q.FindYearlyContestRegistrationForUser(ctx, postgres.FindYearlyContestRegistrationForUserParams{
		UserID:         req.UserID,
		Year:           int32(req.Year),
		IncludePrivate: req.IncludePrivate,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.ContestRegistrations{
				Registrations: []query.ContestRegistration{},
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}
		return nil, fmt.Errorf("could not fetch contest registrations: %w", err)
	}

	languages, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest registrations: %w", err)
	}

	activities, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest registrations: %w", err)
	}

	langs := map[string]string{}
	acts := map[int32]string{}

	for _, l := range languages {
		langs[l.Code] = l.Name
	}
	for _, a := range activities {
		acts[a.ID] = a.Name
	}

	res := &query.ContestRegistrations{
		Registrations: make([]query.ContestRegistration, len(regs)),
		TotalSize:     len(regs),
		NextPageToken: "",
	}
	for i, r := range regs {
		r := r

		// TODO: refactor this out to a mapper
		contest := &query.ContestView{
			ID:                r.ContestID,
			ContestStart:      r.ContestStart,
			ContestEnd:        r.ContestEnd,
			RegistrationEnd:   r.RegistrationEnd,
			Title:             r.Title,
			Description:       postgres.NewStringFromNullString(r.Description),
			Private:           r.Private,
			Official:          r.Official,
			AllowedLanguages:  make([]query.Language, 0),
			AllowedActivities: make([]query.Activity, len(r.ActivityTypeIDAllowList)),
		}

		for i, a := range r.ActivityTypeIDAllowList {
			contest.AllowedActivities[i] = query.Activity{
				ID:   a,
				Name: acts[a],
			}
		}

		reg := query.ContestRegistration{
			ID:              r.ID,
			ContestID:       r.ContestID,
			UserID:          r.UserID,
			UserDisplayName: r.UserDisplayName,
			Languages:       make([]query.Language, len(r.LanguageCodes)),
			Contest:         contest,
		}

		for i, code := range r.LanguageCodes {
			reg.Languages[i] = query.Language{
				Code: code,
				Name: langs[code],
			}
		}

		res.Registrations[i] = reg
	}

	return res, nil
}

func (r *Repository) YearlyActivitySplitForUser(ctx context.Context, req *query.YearlyActivitySplitForUserRequest) (*query.YearlyActivitySplitForUserResponse, error) {
	rows, err := r.q.YearlyActivitySplitForUser(ctx, postgres.YearlyActivitySplitForUserParams{
		UserID: req.UserID,
		Year:   int16(req.Year),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.YearlyActivitySplitForUserResponse{
				Activities: []query.ActivityScore{},
			}, nil
		}
		return nil, fmt.Errorf("could not fetch activity split: %w", err)
	}

	scores := make([]query.ActivityScore, len(rows))
	for i, row := range rows {
		row := row
		scores[i] = query.ActivityScore{
			ActivityID:   int(row.LogActivityID),
			ActivityName: row.LogActivityName,
			Score:        row.Score,
		}
	}

	return &query.YearlyActivitySplitForUserResponse{
		Activities: scores,
	}, nil
}
