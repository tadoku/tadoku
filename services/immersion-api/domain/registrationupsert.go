package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type RegistrationUpsertRepository interface {
	FindContestByID(context.Context, *ContestFindRequest) (*ContestView, error)
	FindRegistrationForUser(context.Context, *RegistrationFindRequest) (*ContestRegistration, error)
	UpsertContestRegistration(context.Context, *RegistrationUpsertRequest) error
	DetachContestLogsForLanguages(context.Context, *DetachContestLogsForLanguagesRequest) error
}

type RegistrationUpsertRequest struct {
	ContestID     uuid.UUID
	LanguageCodes []string

	// Set by domain layer (unexported: only domain can write, others read via getters)
	id              uuid.UUID
	userID          uuid.UUID
	officialContest bool
	year            int16
}

func (r *RegistrationUpsertRequest) ID() uuid.UUID         { return r.id }
func (r *RegistrationUpsertRequest) UserID() uuid.UUID     { return r.userID }
func (r *RegistrationUpsertRequest) OfficialContest() bool { return r.officialContest }
func (r *RegistrationUpsertRequest) Year() int16           { return r.year }

type DetachContestLogsForLanguagesRequest struct {
	ContestID       uuid.UUID
	UserID          uuid.UUID
	LanguageCodes   []string
	OfficialContest bool
	Year            int16
}

type RegistrationUpsert struct {
	repo       RegistrationUpsertRepository
	userUpsert *UserUpsert
}

func NewRegistrationUpsert(
	repo RegistrationUpsertRepository,
	userUpsert *UserUpsert,
) *RegistrationUpsert {
	return &RegistrationUpsert{
		repo:       repo,
		userUpsert: userUpsert,
	}
}

func (s *RegistrationUpsert) Execute(ctx context.Context, req *RegistrationUpsertRequest) error {
	if err := requireAuthentication(ctx); err != nil {
		return err
	}

	if err := s.userUpsert.Execute(ctx); err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	// Enrich request with session
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return ErrUnauthorized
	}
	req.userID = uuid.MustParse(session.Subject)
	req.id = uuid.New()

	contest, err := s.repo.FindContestByID(ctx, &ContestFindRequest{
		ID:             req.ContestID,
		IncludeDeleted: false,
	})
	if err != nil {
		return fmt.Errorf("could not find contest: %w", err)
	}

	if len(req.LanguageCodes) < 1 || len(req.LanguageCodes) > 3 {
		return fmt.Errorf("invalid language code length: %w", ErrInvalidContestRegistration)
	}

	// check if languages are allowed by contest
	if len(contest.AllowedLanguages) > 0 {
		langs := map[string]bool{}
		for _, lang := range contest.AllowedLanguages {
			langs[lang.Code] = true
		}
		for _, code := range req.LanguageCodes {
			if _, ok := langs[code]; !ok {
				return fmt.Errorf("language %s is not allowed by contest: %w", code, ErrInvalidContestRegistration)
			}
		}
	}

	// check if existing registration
	registration, err := s.repo.FindRegistrationForUser(ctx, &RegistrationFindRequest{
		UserID:    req.userID,
		ContestID: req.ContestID,
	})
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	// detach logs for any removed languages
	if registration != nil {
		req.id = registration.ID

		newLangs := map[string]bool{}
		for _, lang := range req.LanguageCodes {
			newLangs[lang] = true
		}

		var removedLanguages []string
		for _, lang := range registration.Languages {
			if _, ok := newLangs[lang.Code]; !ok {
				removedLanguages = append(removedLanguages, lang.Code)
			}
		}

		if len(removedLanguages) > 0 {
			err := s.repo.DetachContestLogsForLanguages(ctx, &DetachContestLogsForLanguagesRequest{
				ContestID:       req.ContestID,
				UserID:          req.userID,
				LanguageCodes:   removedLanguages,
				OfficialContest: contest.Official,
				Year:            int16(contest.ContestStart.Year()),
			})
			if err != nil {
				return fmt.Errorf("could not detach logs for removed languages: %w", err)
			}
		}
	}

	req.officialContest = contest.Official
	req.year = int16(contest.ContestStart.Year())

	if err := s.repo.UpsertContestRegistration(ctx, req); err != nil {
		return err
	}

	return nil
}
