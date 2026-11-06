package logs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestwindow"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/uuidx"
)

func (s *Service) PlanContestRegistrationUpdate(
	log *Log,
	targets []logscore.Target,
	now time.Time,
) ([]logscore.Target, []uuid.UUID, error) {
	desired := make(map[uuid.UUID]struct{}, len(targets))
	for _, target := range targets {
		desired[target.RegistrationID] = struct{}{}
	}
	current := make(map[uuid.UUID]RegistrationReference, len(log.Registrations))
	for _, reference := range log.Registrations {
		if contestwindow.Unfinished(reference.ContestEnd, now) {
			current[reference.RegistrationID] = reference
		}
	}

	toAttach := make([]logscore.Target, 0, len(targets))
	for _, target := range targets {
		if _, exists := current[target.RegistrationID]; !exists {
			toAttach = append(toAttach, target)
		}
	}
	toDetach := make([]uuid.UUID, 0)
	for registrationID, reference := range current {
		if _, exists := desired[registrationID]; !exists {
			toDetach = append(toDetach, reference.ContestID)
		}
	}
	if len(toAttach) > 0 && log.Tracking.Amount == nil && log.Tracking.DurationSeconds == nil {
		return nil, nil, errx.NewInvalidInputError("log tracking data is required for contest attachment")
	}
	return toAttach, toDetach, nil
}

func (s *Service) UpdateContestRegistrations(
	ctx context.Context,
	logID uuid.UUID,
	now time.Time,
	attachments []logscore.ContestTracking,
	detachments []uuid.UUID,
) (MutationResult, error) {
	if _, err := s.lockMutableLog(ctx, logID); err != nil {
		return MutationResult{}, err
	}
	beforeViews, err := s.logs.ContestLogViews(ctx, logID, nil)
	if err != nil {
		return MutationResult{}, err
	}

	for _, contestID := range detachments {
		if err := s.logs.DetachContest(ctx, logID, contestID); err != nil {
			return MutationResult{}, err
		}
	}
	for _, attachment := range attachments {
		if err := s.logs.CreateContestLog(ctx, logID, attachment); err != nil {
			return MutationResult{}, err
		}
	}
	if err := s.recomputeOfficialEligibility(ctx, logID, now); err != nil {
		return MutationResult{}, err
	}

	afterViews, err := s.logs.ContestLogViews(ctx, logID, nil)
	if err != nil {
		return MutationResult{}, err
	}
	return MutationResult{Events: contestLogChanges(logID, beforeViews, afterViews)}, nil
}

func (s *Service) lockMutableLog(ctx context.Context, logID uuid.UUID) (owner uuid.UUID, err error) {
	locked, err := s.logs.LockLog(ctx, logID)
	if err != nil {
		return uuid.Nil, err
	}
	if locked.Frozen {
		return uuid.Nil, ErrLogFrozen
	}
	return locked.UserID, nil
}

// RecomputeOfficialEligibility must run in the transaction that detached the logs from their contests.
func (s *Service) RecomputeOfficialEligibility(ctx context.Context, logIDs []uuid.UUID, now time.Time) error {
	for _, logID := range logIDs {
		if err := s.recomputeOfficialEligibility(ctx, logID, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) recomputeOfficialEligibility(ctx context.Context, logID uuid.UUID, now time.Time) error {
	official, err := s.logs.AttachedContestOfficialFlags(ctx, logID)
	if err != nil {
		return err
	}
	return s.logs.SetOfficialEligibility(ctx, logID, logscore.EligibleOfficial(official), now)
}

func (s *Service) Delete(ctx context.Context, logID uuid.UUID, now time.Time) (MutationResult, error) {
	if _, err := s.lockMutableLog(ctx, logID); err != nil {
		return MutationResult{}, err
	}
	contests, err := s.logs.AttachedContests(ctx, logID)
	if err != nil {
		return MutationResult{}, err
	}
	for _, contest := range contests {
		if !contestwindow.Unfinished(contest.End, now) {
			return MutationResult{}, errx.NewForbiddenError("forbidden")
		}
	}
	if err := s.logs.SoftDelete(ctx, logID, now); err != nil {
		return MutationResult{}, err
	}

	result := MutationResult{}
	for _, contest := range contests {
		result.Events = append(result.Events, contestevent.Change{
			ContestID: contest.ID,
			LogID:     logID,
			Type:      contestevent.Withdrawn,
		})
	}
	return result, nil
}

func (s *Service) ModerateDetach(
	ctx context.Context,
	logID, contestID uuid.UUID,
	now time.Time,
) (MutationResult, error) {
	beforeViews, err := s.logs.ContestLogViews(ctx, logID, nil)
	if err != nil {
		return MutationResult{}, err
	}

	if err := s.logs.DetachContest(ctx, logID, contestID); err != nil {
		return MutationResult{}, err
	}
	if err := s.recomputeOfficialEligibility(ctx, logID, now); err != nil {
		return MutationResult{}, err
	}

	afterViews, err := s.logs.ContestLogViews(ctx, logID, nil)
	if err != nil {
		return MutationResult{}, err
	}
	return MutationResult{Events: contestLogChanges(logID, beforeViews, afterViews)}, nil
}

func (s *Service) Create(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
	description *string,
	scored logscore.Result,
) (CreateResult, error) {
	if err := validateDescription(description); err != nil {
		return CreateResult{}, err
	}

	mutation := logMutation{
		ID:                          uuidx.New(ctx),
		UserID:                      userID,
		LanguageCode:                scored.LanguageCode,
		ActivityID:                  scored.ActivityID,
		Description:                 description,
		Tags:                        scored.Tags,
		Tracking:                    scored.Tracking,
		ContestTrackings:            scored.ContestTrackings,
		EligibleOfficialLeaderboard: scored.EligibleOfficial,
		Year:                        int16(now.Year()),
		Now:                         now,
	}
	created, err := s.create(ctx, mutation)
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{ID: mutation.ID, Events: created.Events}, nil
}

func (s *Service) create(ctx context.Context, mutation logMutation) (MutationResult, error) {
	result := MutationResult{}
	if err := s.logs.CreateLog(ctx, mutation); err != nil {
		return MutationResult{}, err
	}
	for _, tracking := range mutation.ContestTrackings {
		if err := s.logs.CreateContestLog(ctx, mutation.ID, tracking); err != nil {
			return MutationResult{}, err
		}
	}
	for _, tag := range mutation.Tags {
		if err := s.logs.InsertTag(ctx, mutation.ID, mutation.UserID, tag); err != nil {
			return MutationResult{}, err
		}
	}

	seen := make(map[uuid.UUID]struct{}, len(mutation.ContestTrackings))
	for _, tracking := range mutation.ContestTrackings {
		if _, exists := seen[tracking.ContestID]; exists {
			continue
		}
		seen[tracking.ContestID] = struct{}{}
		result.Events = append(result.Events, contestevent.Change{
			ContestID: tracking.ContestID,
			LogID:     mutation.ID,
			Type:      contestevent.Submitted,
		})
	}
	return result, nil
}

func (s *Service) Update(
	ctx context.Context,
	id uuid.UUID,
	now time.Time,
	description *string,
	scored logscore.Result,
) (MutationResult, error) {
	if err := validateDescription(description); err != nil {
		return MutationResult{}, err
	}

	mutation := logMutation{
		ID:               id,
		Description:      description,
		Tags:             scored.Tags,
		Tracking:         scored.Tracking,
		ContestTrackings: scored.ContestTrackings,
		Now:              now,
	}
	return s.update(ctx, mutation)
}

func (s *Service) update(ctx context.Context, mutation logMutation) (MutationResult, error) {
	userID, err := s.lockMutableLog(ctx, mutation.ID)
	if err != nil {
		return MutationResult{}, err
	}
	beforeViews, err := s.logs.ContestLogViews(ctx, mutation.ID, nil)
	if err != nil {
		return MutationResult{}, err
	}

	if err := s.logs.UpdateLog(ctx, mutation); err != nil {
		return MutationResult{}, err
	}
	earliestUnfinishedContestEnd := contestwindow.EarliestUnfinishedContestEnd(mutation.Now)
	if len(mutation.ContestTrackings) == 0 {
		inherited := mutation.Tracking
		inherited.RuleSetID = nil
		inherited.RuleIDs = nil
		inherited.Rates = nil
		inherited.Source = ""
		if err := s.logs.UpdateOngoingContestLogs(ctx, mutation.ID, inherited, earliestUnfinishedContestEnd); err != nil {
			return MutationResult{}, err
		}
	} else {
		for _, tracking := range mutation.ContestTrackings {
			if err := s.logs.UpdateContestLog(ctx, mutation.ID, tracking, earliestUnfinishedContestEnd); err != nil {
				return MutationResult{}, err
			}
		}
	}
	if err := s.logs.DeleteTags(ctx, mutation.ID); err != nil {
		return MutationResult{}, err
	}
	for _, tag := range mutation.Tags {
		if err := s.logs.InsertTag(ctx, mutation.ID, userID, tag); err != nil {
			return MutationResult{}, err
		}
	}

	afterViews, err := s.logs.ContestLogViews(ctx, mutation.ID, nil)
	if err != nil {
		return MutationResult{}, err
	}
	return MutationResult{Events: contestLogChanges(mutation.ID, beforeViews, afterViews)}, nil
}

func (s *Service) RegistrationsForRescoring(log *Log, now time.Time) []logscore.Target {
	if !s.scoringEngineEnabled {
		return nil
	}
	selected := make([]logscore.Target, 0, len(log.Registrations))
	for _, registration := range log.Registrations {
		if contestwindow.Unfinished(registration.ContestEnd, now) {
			selected = append(selected, logscore.Target{
				RegistrationID: registration.RegistrationID,
				ContestID:      registration.ContestID,
			})
		}
	}
	return selected
}
