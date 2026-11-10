package logs

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) ContestProfileSummary(
	ctx context.Context,
	userID, contestID uuid.UUID,
	filters Filters,
) (*ContestProfileSummary, error) {
	if err := filters.Validate(); err != nil {
		return nil, err
	}
	query := filters.query()

	days, err := s.logs.ContestProfileDays(ctx, userID, contestID, query)
	if err != nil {
		return nil, err
	}
	languages, err := s.logs.ContestProfileLanguages(ctx, userID, contestID, query)
	if err != nil {
		return nil, err
	}
	activityScores, err := s.logs.ContestProfileActivities(ctx, userID, contestID, query)
	if err != nil {
		return nil, err
	}
	tags, err := s.logs.ContestProfileTags(ctx, userID, contestID, query)
	if err != nil {
		return nil, err
	}

	for i := range activityScores {
		activity, err := findActivity(int32(activityScores[i].ActivityID))
		if err != nil {
			return nil, err
		}
		activityScores[i].ActivityName = activity.Name
	}

	summary := summarizeDays(days)
	summary.Languages = languages
	summary.Activities = activityScores
	summary.Tags = tags
	return summary, nil
}

// summarizeDays expects days ordered by date.
func summarizeDays(days []ContestDayScore) *ContestProfileSummary {
	summary := &ContestProfileSummary{Days: days}
	totals := []DayScore{}
	for _, day := range days {
		summary.Score += day.Score
		summary.LogCount += day.logCount

		last := len(totals) - 1
		if last < 0 || !totals[last].Date.Equal(day.Date) {
			totals = append(totals, DayScore{Date: day.Date})
			last++
		}
		totals[last].Score += day.Score
	}

	summary.ActiveDays = len(totals)
	for i := range totals {
		if summary.BestDay == nil || totals[i].Score > summary.BestDay.Score {
			summary.BestDay = &totals[i]
		}
	}
	return summary
}
