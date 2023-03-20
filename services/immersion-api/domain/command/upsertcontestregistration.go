package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

type UpsertContestRegistrationRequest struct {
	ID              uuid.UUID
	ContestID       uuid.UUID
	UserID          uuid.UUID
	UserDisplayName string
	LanguageCodes   []string
}

func (s *ServiceImpl) UpsertContestRegistration(ctx context.Context, req *UpsertContestRegistrationRequest) error {
	if domain.IsRole(ctx, domain.RoleBanned) {
		return ErrForbidden
	}
	if domain.IsRole(ctx, domain.RoleGuest) {
		return ErrUnauthorized
	}

	if err := s.UpdateUserMetadataFromSession(ctx); err != nil {
		fmt.Println(err)
		return fmt.Errorf("could not update user: %w", err)
	}

	// Enrich request with session
	session := domain.ParseSession(ctx)
	if session == nil {
		return ErrUnauthorized
	}
	req.UserID = uuid.MustParse(session.Subject)
	req.UserDisplayName = session.DisplayName
	req.ID = uuid.New()

	contest, err := s.r.FindContestByID(ctx, &query.FindContestByIDRequest{
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
	registration, err := s.r.FindRegistrationForUser(ctx, &query.FindRegistrationForUserRequest{
		UserID:    req.UserID,
		ContestID: req.ContestID,
	})
	if err != nil && !errors.Is(err, query.ErrNotFound) {
		return err
	}

	// check if previous languages are included in new set
	if registration != nil {
		req.ID = registration.ID

		langs := map[string]bool{}
		for _, lang := range req.LanguageCodes {
			langs[lang] = true
		}
		for _, lang := range registration.Languages {
			if _, ok := langs[lang.Code]; !ok {
				return fmt.Errorf("language %s is missing but was previously registered: %w", lang.Code, ErrInvalidContestRegistration)
			}
		}
	}

	return s.r.UpsertContestRegistration(ctx, req)
}
