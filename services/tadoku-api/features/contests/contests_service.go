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

	total, err := s.contests.CountContests(ctx, parameters)
	if err != nil {
		return nil, err
	}
	items, err := s.contests.ListContests(ctx, parameters)
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
	item, languages, err := s.findContestWithLanguages(ctx, id, includeDeleted)
	if err != nil {
		return nil, err
	}
	return hydrateContest(item, languages)
}

func (s *Service) FindLatestOfficialContest(ctx context.Context) (*ContestView, error) {
	item, err := s.contests.FindLatestOfficialContest(ctx)
	if err != nil {
		return nil, err
	}
	languages, err := s.languagesForContest(ctx, item)
	if err != nil {
		return nil, err
	}
	return hydrateContest(item, languages)
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

func (s *Service) findContestWithLanguages(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Contest, []Language, error) {
	item, err := s.contests.FindContestByID(ctx, FindParameters{ID: id, includeDeleted: includeDeleted})
	if err != nil {
		return nil, nil, err
	}
	languages, err := s.languagesForContest(ctx, item)
	if err != nil {
		return nil, nil, err
	}
	return item, languages, nil
}

func (s *Service) languagesForContest(ctx context.Context, item *Contest) ([]Language, error) {
	if len(item.LanguageCodeAllowList) == 0 {
		return nil, nil
	}
	return s.contests.ListLanguagesForContest(ctx, item.ID)
}

func hydrateContest(item *Contest, languages []Language) (*ContestView, error) {
	activities, err := hydrateActivities(item.ActivityTypeIDAllowList)
	if err != nil {
		return nil, err
	}
	return &ContestView{
		ID:                   item.ID,
		ContestStart:         item.ContestStart,
		ContestEnd:           item.ContestEnd,
		RegistrationEnd:      item.RegistrationEnd,
		Title:                item.Title,
		Description:          item.Description,
		OwnerUserID:          item.OwnerUserID,
		OwnerUserDisplayName: item.OwnerUserDisplayName,
		Official:             item.Official,
		Private:              item.Private,
		AllowedLanguages:     languages,
		AllowedActivities:    activities,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
		Deleted:              item.Deleted,
	}, nil
}
