package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogContestUpdateRepository interface {
	FetchOngoingContestRegistrations(context.Context, *RegistrationListOngoingRequest) (*ContestRegistrations, error)
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
	UpdateLogContests(ctx context.Context, req *LogContestUpdateDBRequest) error
}

type LogContestUpdateRequest struct {
	LogID           uuid.UUID
	RegistrationIDs []uuid.UUID
}

type LogContestUpdateDBRequest struct {
	LogID    uuid.UUID
	Amount   float32
	Modifier float32
	Attach   []LogContestAttach
	Detach   []uuid.UUID // contest_ids to remove
}

type LogContestAttach struct {
	RegistrationID uuid.UUID
	ContestID      uuid.UUID
}

type LogContestUpdate struct {
	repo  LogContestUpdateRepository
	clock commondomain.Clock
}

func NewLogContestUpdate(
	repo LogContestUpdateRepository,
	clock commondomain.Clock,
) *LogContestUpdate {
	return &LogContestUpdate{
		repo:  repo,
		clock: clock,
	}
}

func (s *LogContestUpdate) Execute(ctx context.Context, req *LogContestUpdateRequest) (*Log, error) {
	if err := requireAuthentication(ctx); err != nil {
		return nil, err
	}

	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userID := uuid.MustParse(session.Subject)

	// Fetch the log to verify ownership and get current state
	log, err := s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not find log: %w", err)
	}

	if log.UserID != userID {
		return nil, ErrForbidden
	}

	now := s.clock.Now()

	// Fetch ongoing registrations to validate requested IDs
	registrations, err := s.repo.FetchOngoingContestRegistrations(ctx, &RegistrationListOngoingRequest{
		UserID: userID,
		Now:    now,
	})
	if err != nil {
		return nil, fmt.Errorf("could not fetch registrations: %w", err)
	}

	// Build lookup: registration_id → ContestRegistration
	validRegistrations := map[uuid.UUID]ContestRegistration{}
	for _, r := range registrations.Registrations {
		validRegistrations[r.ID] = r
	}

	// Validate each requested registration
	for _, regID := range req.RegistrationIDs {
		registration, ok := validRegistrations[regID]
		if !ok {
			return nil, fmt.Errorf("registration %s is not an ongoing registration: %w", regID, ErrInvalidLog)
		}

		// Validate language matches
		found := false
		for _, lang := range registration.Languages {
			if lang.Code == log.LanguageCode {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("language %s is not allowed for registration %s: %w", log.LanguageCode, regID, ErrInvalidLog)
		}

		// Validate activity matches
		found = false
		for _, act := range registration.Contest.AllowedActivities {
			if int(act.ID) == log.ActivityID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("activity %d is not allowed for registration %s: %w", log.ActivityID, regID, ErrInvalidLog)
		}
	}

	// Build set of desired registration IDs
	desiredSet := map[uuid.UUID]bool{}
	for _, regID := range req.RegistrationIDs {
		desiredSet[regID] = true
	}

	// Build set of current ongoing registration IDs from the log's attachments
	// Only consider ongoing contests (contest_end + 1 day > now)
	currentOngoing := map[uuid.UUID]ContestRegistrationReference{}
	for _, ref := range log.Registrations {
		if ref.ContestEnd.Add(24 * time.Hour).After(now) {
			currentOngoing[ref.RegistrationID] = ref
		}
	}

	// Compute diff
	var toAttach []LogContestAttach
	var toDetach []uuid.UUID

	for _, regID := range req.RegistrationIDs {
		if _, alreadyAttached := currentOngoing[regID]; !alreadyAttached {
			reg := validRegistrations[regID]
			toAttach = append(toAttach, LogContestAttach{
				RegistrationID: regID,
				ContestID:      reg.ContestID,
			})
		}
	}

	for regID, ref := range currentOngoing {
		if !desiredSet[regID] {
			toDetach = append(toDetach, ref.ContestID)
		}
	}

	if len(toAttach) == 0 && len(toDetach) == 0 {
		return log, nil
	}

	if err := s.repo.UpdateLogContests(ctx, &LogContestUpdateDBRequest{
		LogID:    req.LogID,
		Amount:   log.Amount,
		Modifier: log.Modifier,
		Attach:   toAttach,
		Detach:   toDetach,
	}); err != nil {
		return nil, fmt.Errorf("could not update log contests: %w", err)
	}

	updatedLog, err := s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not fetch updated log: %w", err)
	}

	return updatedLog, nil
}
