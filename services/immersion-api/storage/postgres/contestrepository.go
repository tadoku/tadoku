package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/contestcommand"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestquery"
	"github.com/tadoku/tadoku/services/immersion-api/domain/profilequery"
)

type ContestRepository struct {
	psql *sql.DB
	q    *Queries
}

func NewContestRepository(psql *sql.DB) *ContestRepository {
	return &ContestRepository{
		psql: psql,
		q:    &Queries{psql},
	}
}

// COMMANDS

func (r *ContestRepository) CreateContest(ctx context.Context, req *contestcommand.ContestCreateRequest) (*contestcommand.ContestCreateResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create contest: %w", err)
	}

	qtx := r.q.WithTx(tx)

	id, err := qtx.CreateContest(ctx, CreateContestParams{
		OwnerUserID:             req.OwnerUserID,
		OwnerUserDisplayName:    req.OwnerUserDisplayName,
		Official:                req.Official,
		Private:                 req.Private,
		ContestStart:            req.ContestStart,
		ContestEnd:              req.ContestEnd,
		RegistrationEnd:         req.RegistrationEnd,
		Title:                   req.Title,
		Description:             NewNullString(req.Description),
		LanguageCodeAllowList:   req.LanguageCodeAllowList,
		ActivityTypeIDAllowList: req.ActivityTypeIDAllowList,
	})

	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create contest: %w", err)
	}

	contest, err := qtx.FindContestById(ctx, FindContestByIdParams{
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

	return &contestcommand.ContestCreateResponse{
		ID:                      contest.ID,
		ContestStart:            contest.ContestStart,
		ContestEnd:              contest.ContestEnd,
		RegistrationEnd:         contest.RegistrationEnd,
		Title:                   contest.Title,
		Description:             NewStringFromNullString(contest.Description),
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

func (r *ContestRepository) UpsertContestRegistration(ctx context.Context, req *contestcommand.UpsertContestRegistrationRequest) error {
	_, err := r.q.UpsertContestRegistration(ctx, UpsertContestRegistrationParams{
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

func (r *ContestRepository) FetchContestConfigurationOptions(ctx context.Context) (*contestquery.FetchContestConfigurationOptionsResponse, error) {
	langs, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest configuration options: %w", err)
	}

	acts, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest configuration options: %w", err)
	}

	options := contestquery.FetchContestConfigurationOptionsResponse{
		Languages:  make([]contestquery.Language, len(langs)),
		Activities: make([]contestquery.Activity, len(acts)),
	}

	for i, l := range langs {
		options.Languages[i] = contestquery.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	for i, a := range acts {
		options.Activities[i] = contestquery.Activity{
			ID:      a.ID,
			Name:    a.Name,
			Default: a.Default,
		}
	}

	return &options, err
}

func (r *ContestRepository) FindLatestOfficial(ctx context.Context) (*contestquery.ContestView, error) {
	contest, err := r.q.FindLatestOfficialContest(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, contestquery.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	activities, err := r.q.ListActivitiesForContest(ctx, contest.ID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	acts := make([]contestquery.Activity, len(activities))
	for i, a := range activities {
		acts[i] = contestquery.Activity{
			ID:   a.ID,
			Name: a.Name,
		}
	}

	langs := []contestquery.Language{}

	if len(contest.LanguageCodeAllowList) > 0 {
		languages, err := r.q.ListLanguagesForContest(ctx, contest.ID)
		if err != nil {
			return nil, fmt.Errorf("could not fetch contest: %w", err)
		}

		langs = make([]contestquery.Language, len(languages))
		for i, a := range languages {
			langs[i] = contestquery.Language{
				Code: a.Code,
				Name: a.Name,
			}
		}
	}

	return &contestquery.ContestView{
		ID:                   contest.ID,
		ContestStart:         contest.ContestStart,
		ContestEnd:           contest.ContestEnd,
		RegistrationEnd:      contest.RegistrationEnd,
		Title:                contest.Title,
		Description:          NewStringFromNullString(contest.Description),
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

func (r *ContestRepository) FindByID(ctx context.Context, req *contestquery.FindByIDRequest) (*contestquery.ContestView, error) {
	contest, err := r.q.FindContestById(ctx, FindContestByIdParams{
		ID:             req.ID,
		IncludeDeleted: req.IncludeDeleted,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, contestquery.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	activities, err := r.q.ListActivitiesForContest(ctx, contest.ID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	acts := make([]contestquery.Activity, len(activities))
	for i, a := range activities {
		acts[i] = contestquery.Activity{
			ID:   a.ID,
			Name: a.Name,
		}
	}

	langs := []contestquery.Language{}

	if len(contest.LanguageCodeAllowList) > 0 {
		languages, err := r.q.ListLanguagesForContest(ctx, contest.ID)
		if err != nil {
			return nil, fmt.Errorf("could not fetch contest: %w", err)
		}

		langs = make([]contestquery.Language, len(languages))
		for i, a := range languages {
			langs[i] = contestquery.Language{
				Code: a.Code,
				Name: a.Name,
			}
		}
	}

	return &contestquery.ContestView{
		ID:                   contest.ID,
		ContestStart:         contest.ContestStart,
		ContestEnd:           contest.ContestEnd,
		RegistrationEnd:      contest.RegistrationEnd,
		Title:                contest.Title,
		Description:          NewStringFromNullString(contest.Description),
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

func (r *ContestRepository) ListContests(ctx context.Context, req *contestquery.ContestListRequest) (*contestquery.ContestListResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list contests: %w", err)
	}

	qtx := r.q.WithTx(tx)

	meta, err := qtx.ContestsMetadata(ctx, ContestsMetadataParams{
		IncludeDeleted: req.IncludeDeleted,
		UserID:         req.UserID,
		Official:       req.OfficialOnly,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not lists contests: %w", err)
	}

	contests, err := qtx.ListContests(ctx, ListContestsParams{
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

	res := make([]contestquery.Contest, len(contests))
	for i, c := range contests {
		res[i] = contestquery.Contest{
			ID:                      c.ID,
			ContestStart:            c.ContestStart,
			ContestEnd:              c.ContestEnd,
			RegistrationEnd:         c.RegistrationEnd,
			Title:                   c.Title,
			Description:             NewStringFromNullString(c.Description),
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

	return &contestquery.ContestListResponse{
		Contests:      res,
		TotalSize:     int(meta.TotalSize),
		NextPageToken: nextPageToken,
	}, nil
}

func (r *ContestRepository) FindRegistrationForUser(ctx context.Context, req *contestquery.FindRegistrationForUserRequest) (*contestquery.ContestRegistration, error) {
	reg, err := r.q.FindContestRegistrationForUser(ctx, FindContestRegistrationForUserParams{
		UserID:    req.UserID,
		ContestID: req.ContestID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, contestquery.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch contest registration: %w", err)
	}

	languages, err := r.q.GetLanguagesByCode(ctx, reg.LanguageCodes)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest registration: %w", err)
	}

	langs := make([]contestquery.Language, len(languages))
	for i, a := range languages {
		langs[i] = contestquery.Language{
			Code: a.Code,
			Name: a.Name,
		}
	}

	return &contestquery.ContestRegistration{
		ID:              reg.ID,
		ContestID:       reg.ContestID,
		UserID:          req.UserID,
		UserDisplayName: reg.UserDisplayName,
		Languages:       langs,
	}, nil
}

func (r *ContestRepository) FetchContestLeaderboard(ctx context.Context, req *contestquery.FetchContestLeaderboardRequest) (*contestquery.Leaderboard, error) {
	_, err := r.q.FindContestById(ctx, FindContestByIdParams{ID: req.ContestID, IncludeDeleted: false})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, contestquery.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch leaderboard for contest: %w", err)
	}

	entries, err := r.q.LeaderboardForContest(ctx, LeaderboardForContestParams{
		ContestID:    req.ContestID,
		LanguageCode: NewNullString(req.LanguageCode),
		ActivityID:   NewNullInt32(req.ActivityID),
		StartFrom:    int32(req.Page * req.PageSize),
		PageSize:     int32(req.PageSize),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &contestquery.Leaderboard{
				Entries:       []contestquery.LeaderboardEntry{},
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch leaderboard for contest: %w", err)
	}

	res := make([]contestquery.LeaderboardEntry, len(entries))
	for i, e := range entries {
		res[i] = contestquery.LeaderboardEntry{
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

	return &contestquery.Leaderboard{
		Entries:       res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

func (r *ContestRepository) FetchOngoingContestRegistrations(ctx context.Context, req *contestquery.FetchOngoingContestRegistrationsRequest) (*contestquery.ContestRegistrations, error) {
	regs, err := r.q.FindOngoingContestRegistrationForUser(ctx, FindOngoingContestRegistrationForUserParams{
		UserID: req.UserID,
		Now:    req.Now,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &contestquery.ContestRegistrations{
				Registrations: []contestquery.ContestRegistration{},
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

	res := &contestquery.ContestRegistrations{
		Registrations: make([]contestquery.ContestRegistration, len(regs)),
		TotalSize:     len(regs),
		NextPageToken: "",
	}
	for i, r := range regs {
		r := r

		contest := &contestquery.ContestView{
			ID:                r.ContestID,
			ContestStart:      r.ContestStart,
			ContestEnd:        r.ContestEnd,
			RegistrationEnd:   r.RegistrationEnd,
			Title:             r.Title,
			Description:       NewStringFromNullString(r.Description),
			Private:           r.Private,
			AllowedLanguages:  make([]contestquery.Language, 0),
			AllowedActivities: make([]contestquery.Activity, len(r.ActivityTypeIDAllowList)),
		}

		for i, a := range r.ActivityTypeIDAllowList {
			contest.AllowedActivities[i] = contestquery.Activity{
				ID:   a,
				Name: acts[a],
			}
		}

		reg := contestquery.ContestRegistration{
			ID:              r.ID,
			ContestID:       r.ContestID,
			UserID:          r.UserID,
			UserDisplayName: r.UserDisplayName,
			Languages:       make([]contestquery.Language, len(r.LanguageCodes)),
			Contest:         contest,
		}

		for i, code := range r.LanguageCodes {
			reg.Languages[i] = contestquery.Language{
				Code: code,
				Name: langs[code],
			}
		}

		res.Registrations[i] = reg
	}

	return res, nil
}

func (r *ContestRepository) FindScoresForRegistration(ctx context.Context, req *profilequery.ContestProfileRequest) ([]profilequery.Score, error) {
	rows, err := r.q.FetchScoresForContestProfile(ctx, FetchScoresForContestProfileParams{
		ContestID: req.ContestID,
		UserID:    req.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, profilequery.ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	scores := make([]profilequery.Score, len(rows))
	for i, row := range rows {
		scores[i] = profilequery.Score{
			LanguageCode: row.LanguageCode,
			Score:        row.Score,
		}
	}

	return scores, nil
}
