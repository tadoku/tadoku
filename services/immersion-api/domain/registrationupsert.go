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
	ID            uuid.UUID
	ContestID     uuid.UUID
	UserID        uuid.UUID
	LanguageCodes []string
}

type DetachContestLogsForLanguagesRequest struct {
	ContestID     uuid.UUID
	UserID        uuid.UUID
	LanguageCodes []string
}

type RegistrationUpsert struct {
	repo               RegistrationUpsertRepository
	userUpsert         *UserUpsert
	leaderboardUpdater *LeaderboardUpdater
}

func NewRegistrationUpsert(
	repo RegistrationUpsertRepository,
	userUpsert *UserUpsert,
	leaderboardUpdater *LeaderboardUpdater,
) *RegistrationUpsert {
	return &RegistrationUpsert{
		repo:               repo,
		userUpsert:         userUpsert,
		leaderboardUpdater: leaderboardUpdater,
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
	req.UserID = uuid.MustParse(session.Subject)
	req.ID = uuid.New()

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
		UserID:    req.UserID,
		ContestID: req.ContestID,
	})
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	// detach logs for any removed languages
	if registration != nil {
		req.ID = registration.ID

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
				ContestID:     req.ContestID,
				UserID:        req.UserID,
				LanguageCodes: removedLanguages,
			})
			if err != nil {
				return fmt.Errorf("could not detach logs for removed languages: %w", err)
			}
		}
	}

	if err := s.repo.UpsertContestRegistration(ctx, req); err != nil {
		return err
	}

	// Update the user's contest score — best effort, do not fail the registration
	s.leaderboardUpdater.UpdateUserContestScore(ctx, req.ContestID, req.UserID)

	return nil
}
