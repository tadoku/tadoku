package contests

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	kratosapi "github.com/ory/kratos-client-go"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

const contestCreationYearlyLimit = 12

type Service struct {
	contests *ContestsRepository
	kratos   *kratosapi.APIClient
}

func NewService(repository *ContestsRepository, kratos *kratosapi.APIClient) *Service {
	return &Service{
		contests: repository,
		kratos:   kratos,
	}
}

func (s *Service) ValidateContestCreation(
	ctx context.Context,
	parameters CreateContestParameters,
	creatorID uuid.UUID,
	creatorDisplayName string,
	admin bool,
	now time.Time,
) error {
	if !admin {
		count, err := s.contests.CountContestsCreatedByUserForYear(ctx, creatorID, int32(now.Year()))
		if err != nil {
			return err
		}
		if count >= contestCreationYearlyLimit {
			return ErrContestCreationForbidden
		}
	}
	if err := parameters.validate(creatorID, creatorDisplayName, admin, now); err != nil {
		return err
	}
	if len(parameters.LanguageCodeAllowList) > 0 {
		exists, err := s.contests.LanguagesExist(ctx, parameters.LanguageCodeAllowList)
		if err != nil {
			return err
		}
		if !exists {
			return errx.NewInvalidInputError("invalid contest LanguageCodeAllowList: one or more languages do not exist")
		}
	}
	return nil
}

func (s *Service) FindRegistration(ctx context.Context, userID, contestID uuid.UUID) (*Registration, error) {
	registration, err := s.contests.FindRegistrationForUser(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}
	registration.Languages, err = s.contests.ListRegistrationLanguages(ctx, registration.LanguageCodes)
	if err != nil {
		return nil, err
	}
	return registration, nil
}

func (s *Service) ListOngoingRegistrations(ctx context.Context, userID uuid.UUID) (*RegistrationList, error) {
	registrations, err := s.contests.ListOngoingRegistrations(ctx, userID, timex.Now())
	if err != nil {
		return nil, err
	}
	languages, err := s.contests.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}
	languageNames := make(map[string]string, len(languages))
	for _, language := range languages {
		languageNames[language.Code] = language.Name
	}
	for i := range registrations {
		registrations[i].Languages = make([]Language, 0, len(registrations[i].LanguageCodes))
		for _, code := range registrations[i].LanguageCodes {
			registrations[i].Languages = append(registrations[i].Languages, Language{Code: code, Name: languageNames[code]})
		}
		activities, err := hydrateActivitiesInOrder(registrations[i].Contest.allowedActivityIDs)
		if err != nil {
			return nil, err
		}
		registrations[i].Contest.AllowedActivities = activities
	}
	return &RegistrationList{
		Registrations: registrations,
		TotalSize:     len(registrations),
		NextPageToken: "",
	}, nil
}

func (s *Service) ValidateRegistrationUpsert(
	ctx context.Context,
	parameters RegistrationUpsertParameters,
) (*Contest, error) {
	contest, allowedLanguages, err := s.findContestWithLanguages(ctx, parameters.ContestID, false)
	if err != nil {
		return nil, err
	}
	if len(parameters.LanguageCodes) < 1 || len(parameters.LanguageCodes) > 3 {
		return nil, ErrInvalidRegistration
	}
	languages := make(map[string]struct{}, len(parameters.LanguageCodes))
	for _, code := range parameters.LanguageCodes {
		if _, exists := languages[code]; exists {
			return nil, ErrInvalidRegistration
		}
		languages[code] = struct{}{}
	}
	exist, err := s.contests.LanguagesExist(ctx, parameters.LanguageCodes)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ErrInvalidRegistration
	}
	if len(allowedLanguages) > 0 {
		allowed := make(map[string]struct{}, len(allowedLanguages))
		for _, language := range allowedLanguages {
			allowed[language.Code] = struct{}{}
		}
		for _, code := range parameters.LanguageCodes {
			if _, ok := allowed[code]; !ok {
				return nil, ErrInvalidRegistration
			}
		}
	}
	return contest, nil
}

func (s *Service) ApplyRegistration(
	ctx context.Context,
	registration Registration,
	removedLanguages []string,
	contest Contest,
) error {
	if len(removedLanguages) > 0 {
		if err := s.contests.DetachContestLogsForLanguages(
			ctx,
			registration.UserID,
			registration.ContestID,
			removedLanguages,
		); err != nil {
			return err
		}
		if err := s.insertRegistrationLeaderboardOutbox(ctx, registration, contest); err != nil {
			return err
		}
	}
	if err := s.contests.UpsertRegistration(ctx, registration); err != nil {
		return err
	}
	return s.insertRegistrationLeaderboardOutbox(ctx, registration, contest)
}

func (s *Service) insertRegistrationLeaderboardOutbox(ctx context.Context, registration Registration, contest Contest) error {
	if err := s.contests.InsertContestScoreRefresh(ctx, registration.UserID, registration.ContestID); err != nil {
		return err
	}
	if contest.Official {
		return s.contests.InsertOfficialScoresRefresh(ctx, registration.UserID, int16(contest.ContestStart.Year()))
	}
	return nil
}

func (s *Service) CreateContest(ctx context.Context, contest Contest) (*Contest, error) {
	if err := s.contests.CreateContest(ctx, contest); err != nil {
		return nil, err
	}
	return s.contests.FindCreatedContestByID(ctx, contest.ID)
}

func (s *Service) CheckCreatePermission(ctx context.Context, userID uuid.UUID) error {
	identity, response, err := s.kratos.IdentityApi.GetIdentity(ctx, userID.String()).Execute()
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			return ErrContestCreatorNotFound
		}
		return fmt.Errorf("fetch contest creator: %w", err)
	}
	if identity.GetSchemaId() != "user" {
		return fmt.Errorf("unexpected contest creator schema %s", identity.GetSchemaId())
	}

	now := timex.Now()
	if identity.GetCreatedAt().After(now.AddDate(0, -1, 0)) {
		return ErrContestCreatorTooYoung
	}

	count, err := s.contests.CountContestsCreatedByUserForYear(ctx, userID, int32(now.Year()))
	if err != nil {
		return err
	}
	if count >= contestCreationYearlyLimit {
		return ErrContestCreationForbidden
	}
	return nil
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
