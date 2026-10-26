package logs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
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
		if reference.ContestEnd.Add(24 * time.Hour).After(now) {
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
	result := MutationResult{}
	if err := s.logs.LockLog(ctx, logID); err != nil {
		return MutationResult{}, err
	}
	beforeViews, err := s.logs.ContestLogViews(ctx, logID, nil)
	if err != nil {
		return MutationResult{}, err
	}

	before, err := s.logs.JobContext(ctx, logID)
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
	if err := s.logs.RecomputeOfficialEligibility(ctx, logID, now); err != nil {
		return MutationResult{}, err
	}
	after, err := s.logs.JobContext(ctx, logID)
	if err != nil {
		return MutationResult{}, err
	}

	affected := make(map[uuid.UUID]struct{}, len(detachments)+len(attachments))
	for _, contestID := range detachments {
		affected[contestID] = struct{}{}
	}
	for _, attachment := range attachments {
		affected[attachment.ContestID] = struct{}{}
	}
	for contestID := range affected {
		result.Jobs = append(result.Jobs, jobs.InvalidateContestLeaderboardV1{ContestID: contestID})
	}
	if before.EligibleOfficial || after.EligibleOfficial {
		year := before.Year
		result.Jobs = append(result.Jobs, jobs.InvalidateOfficialLeaderboardV1{Year: year})
	}
	afterViews, err := s.logs.ContestLogViews(ctx, logID, nil)
	if err != nil {
		return MutationResult{}, err
	}
	result.Events = contestLogChanges(logID, beforeViews, afterViews)
	return result, nil
}

func (s *Service) Delete(ctx context.Context, logID uuid.UUID, now time.Time) (MutationResult, error) {
	result := MutationResult{}
	if err := s.logs.LockLog(ctx, logID); err != nil {
		return MutationResult{}, err
	}
	allowed, err := s.logs.CanDelete(ctx, logID, now)
	if err != nil {
		return MutationResult{}, err
	}
	if !allowed {
		return MutationResult{}, errx.NewForbiddenError("forbidden")
	}
	jobContext, err := s.logs.JobContext(ctx, logID)
	if err != nil {
		return MutationResult{}, err
	}
	contestIDs, err := s.logs.AttachedContestIDs(ctx, logID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.logs.SoftDelete(ctx, logID, now); err != nil {
		return MutationResult{}, err
	}
	for _, contestID := range contestIDs {
		result.Jobs = append(result.Jobs, jobs.InvalidateContestLeaderboardV1{ContestID: contestID})
		result.Events = append(result.Events, contestevent.Change{
			ContestID: contestID,
			LogID:     logID,
			Type:      contestevent.Withdrawn,
		})
	}
	if jobContext.EligibleOfficial {
		year := jobContext.Year
		result.Jobs = append(result.Jobs, jobs.InvalidateOfficialLeaderboardV1{Year: year})
	}
	return result, nil
}

func (s *Service) ModerateDetach(
	ctx context.Context,
	logID, contestID uuid.UUID,
	now time.Time,
) (MutationResult, error) {
	result := MutationResult{}
	beforeViews, err := s.logs.ContestLogViews(ctx, logID, nil)
	if err != nil {
		return MutationResult{}, err
	}

	before, err := s.logs.JobContext(ctx, logID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.logs.DetachContest(ctx, logID, contestID); err != nil {
		return MutationResult{}, err
	}
	if err := s.logs.RecomputeOfficialEligibility(ctx, logID, now); err != nil {
		return MutationResult{}, err
	}
	after, err := s.logs.JobContext(ctx, logID)
	if err != nil {
		return MutationResult{}, err
	}

	result.Jobs = append(result.Jobs, jobs.InvalidateContestLeaderboardV1{ContestID: contestID})
	if before.EligibleOfficial || after.EligibleOfficial {
		year := before.Year
		result.Jobs = append(result.Jobs, jobs.InvalidateOfficialLeaderboardV1{Year: year})
	}
	afterViews, err := s.logs.ContestLogViews(ctx, logID, nil)
	if err != nil {
		return MutationResult{}, err
	}
	result.Events = contestLogChanges(logID, beforeViews, afterViews)
	return result, nil
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
		ID:                          uuid.New(),
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
	followUp, err := s.create(ctx, mutation)
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{ID: mutation.ID, Jobs: followUp.Jobs, Events: followUp.Events}, nil
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
		contestID := tracking.ContestID
		result.Jobs = append(result.Jobs, jobs.InvalidateContestLeaderboardV1{ContestID: contestID})
		result.Events = append(result.Events, contestevent.Change{
			ContestID: contestID,
			LogID:     mutation.ID,
			Type:      contestevent.Submitted,
		})
	}
	if mutation.EligibleOfficialLeaderboard {
		year := mutation.Year
		result.Jobs = append(result.Jobs, jobs.InvalidateOfficialLeaderboardV1{Year: year})
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
	result := MutationResult{}
	if err := s.logs.LockLog(ctx, mutation.ID); err != nil {
		return MutationResult{}, err
	}
	beforeViews, err := s.logs.ContestLogViews(ctx, mutation.ID, nil)
	if err != nil {
		return MutationResult{}, err
	}

	jobContext, err := s.logs.JobContext(ctx, mutation.ID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.logs.UpdateLog(ctx, mutation); err != nil {
		return MutationResult{}, err
	}
	if len(mutation.ContestTrackings) == 0 {
		inherited := mutation.Tracking
		inherited.RuleSetID = nil
		inherited.RuleIDs = nil
		inherited.Rates = nil
		inherited.Source = ""
		if err := s.logs.UpdateOngoingContestLogs(ctx, mutation.ID, inherited, mutation.Now); err != nil {
			return MutationResult{}, err
		}
	} else {
		for _, tracking := range mutation.ContestTrackings {
			if err := s.logs.UpdateContestLog(ctx, mutation.ID, tracking, mutation.Now); err != nil {
				return MutationResult{}, err
			}
		}
	}
	if err := s.logs.DeleteTags(ctx, mutation.ID); err != nil {
		return MutationResult{}, err
	}
	for _, tag := range mutation.Tags {
		if err := s.logs.InsertTag(ctx, mutation.ID, jobContext.UserID, tag); err != nil {
			return MutationResult{}, err
		}
	}
	contestIDs, err := s.logs.OngoingContestIDs(ctx, mutation.ID, mutation.Now)
	if err != nil {
		return MutationResult{}, err
	}
	for _, id := range contestIDs {
		result.Jobs = append(result.Jobs, jobs.InvalidateContestLeaderboardV1{ContestID: id})
	}
	if jobContext.EligibleOfficial {
		year := jobContext.Year
		result.Jobs = append(result.Jobs, jobs.InvalidateOfficialLeaderboardV1{Year: year})
	}
	afterViews, err := s.logs.ContestLogViews(ctx, mutation.ID, nil)
	if err != nil {
		return MutationResult{}, err
	}
	result.Events = contestLogChanges(mutation.ID, beforeViews, afterViews)
	return result, nil
}

func (s *Service) RegistrationsForRescoring(log *Log, now time.Time) []logscore.Target {
	if !s.scoringEngineEnabled {
		return nil
	}
	selected := make([]logscore.Target, 0, len(log.Registrations))
	for _, registration := range log.Registrations {
		if registration.ContestEnd.Add(24 * time.Hour).After(now) {
			selected = append(selected, logscore.Target{
				RegistrationID: registration.RegistrationID,
				ContestID:      registration.ContestID,
			})
		}
	}
	return selected
}
