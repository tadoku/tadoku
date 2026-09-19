package contests

import (
	"context"
	"strconv"

	"github.com/google/uuid"
)

type Service struct {
	contests *ContestsRepository
}

func NewService(repository *ContestsRepository) *Service {
	return &Service{contests: repository}
}

func (s *Service) ListContests(ctx context.Context, parameters ListParameters, includePrivate bool) (*ContestList, error) {
	if parameters.PageSize == 0 {
		parameters.PageSize = 10
	}
	if parameters.PageSize > 100 || parameters.PageSize == 0 {
		parameters.PageSize = 100
	}
	parameters.includePrivate = includePrivate

	items, total, err := s.contests.ListContests(ctx, parameters)
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if parameters.Page*parameters.PageSize+parameters.PageSize < total {
		nextPageToken = strconv.Itoa(parameters.Page + 1)
	}
	return &ContestList{
		Contests:      items,
		TotalSize:     total,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Service) FindContestByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*ContestView, error) {
	item, err := s.contests.FindContestByID(ctx, FindParameters{ID: id, includeDeleted: includeDeleted})
	if err != nil {
		return nil, err
	}
	return hydrateContest(item)
}

func (s *Service) FindLatestOfficialContest(ctx context.Context) (*ContestView, error) {
	item, err := s.contests.FindLatestOfficialContest(ctx)
	if err != nil {
		return nil, err
	}
	return hydrateContest(item)
}

func (s *Service) ConfigurationOptions(ctx context.Context, canCreateOfficialRound bool) (*ConfigurationOptions, error) {
	languages, err := s.contests.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}
	return &ConfigurationOptions{
		Languages:              languages,
		Activities:             allActivities(),
		CanCreateOfficialRound: canCreateOfficialRound,
	}, nil
}

func hydrateContest(item *ContestView) (*ContestView, error) {
	activities, err := hydrateActivities(item.allowedActivityIDs)
	if err != nil {
		return nil, err
	}
	item.AllowedActivities = activities
	return item, nil
}
