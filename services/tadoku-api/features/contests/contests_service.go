package contests

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	domainlanguages "github.com/tadoku/tadoku/services/tadoku-api/domain/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Service struct {
	contests *ContestsRepository
}

func NewService(repository *ContestsRepository) *Service {
	return &Service{
		contests: repository,
	}
}

func (s *Service) FetchContestSummary(ctx context.Context, contestID uuid.UUID) (*ContestSummary, error) {
	return s.contests.FetchContestSummary(ctx, contestID)
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
		if err := checkContestCreationYearlyLimit(count); err != nil {
			return err
		}
	}
	return parameters.validate(creatorID, creatorDisplayName, admin, now)
}

func (s *Service) FindRegistration(ctx context.Context, userID, contestID uuid.UUID, languages []domainlanguages.Language) (*Registration, error) {
	registration, err := s.contests.FindRegistrationForUser(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}

	registration.Languages = registrationLanguages(registration.LanguageCodes, languageNames(languages))
	return registration, nil
}

func (s *Service) FindRegistrationWithContest(ctx context.Context, userID, contestID uuid.UUID, languages []domainlanguages.Language) (*Registration, error) {
	registration, err := s.contests.FindRegistrationWithContestForUser(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}

	registration.Languages = registrationLanguages(registration.LanguageCodes, languageNames(languages))
	return registration, nil
}

func (s *Service) ListOngoingRegistrations(ctx context.Context, userID uuid.UUID, languages []domainlanguages.Language) (*RegistrationList, error) {
	registrations, err := s.contests.ListOngoingRegistrations(ctx, userID, timex.Now())
	if err != nil {
		return nil, err
	}

	return hydrateRegistrations(registrations, languages)
}

func (s *Service) SelectRegistrationsForScoring(ctx context.Context, userID uuid.UUID, requested []uuid.UUID, languageCode string, activityID int32) ([]logscore.Target, error) {
	registrations, err := s.contests.ListOngoingRegistrations(ctx, userID, timex.Now())
	if err != nil {
		return nil, err
	}
	if err := hydrateRegistrationActivities(registrations); err != nil {
		return nil, err
	}

	selected, err := selectRegistrationsForScoring(requested, registrations, languageCode, activityID)
	if err != nil {
		return nil, err
	}
	targets := make([]logscore.Target, 0, len(selected))
	for _, registration := range selected {
		targets = append(targets, logscore.Target{
			RegistrationID: registration.ID,
			ContestID:      registration.ContestID,
			Official:       registration.Contest.Official,
		})
	}
	return targets, nil
}

func (s *Service) ListYearlyRegistrations(ctx context.Context, userID uuid.UUID, year int, includePrivate bool, languages []domainlanguages.Language) (*RegistrationList, error) {
	registrations, err := s.contests.ListYearlyRegistrations(ctx, userID, int32(year), includePrivate)
	if err != nil {
		return nil, err
	}

	return hydrateRegistrations(registrations, languages)
}

func hydrateRegistrations(registrations []Registration, languages []domainlanguages.Language) (*RegistrationList, error) {
	if err := hydrateRegistrationActivities(registrations); err != nil {
		return nil, err
	}

	names := languageNames(languages)
	for i := range registrations {
		registrations[i].Languages = registrationLanguages(registrations[i].LanguageCodes, names)
	}

	return &RegistrationList{
		Registrations: registrations,
		TotalSize:     len(registrations),
		NextPageToken: "",
	}, nil
}

func languageNames(languages []domainlanguages.Language) map[string]string {
	names := make(map[string]string, len(languages))
	for _, language := range languages {
		names[language.Code] = language.Name
	}
	return names
}

// registrationLanguages keeps one language per stored code, in stored order.
// Codes missing from the catalog keep an empty name.
func registrationLanguages(codes []string, names map[string]string) []Language {
	languages := make([]Language, 0, len(codes))
	for _, code := range codes {
		languages = append(languages, Language{Code: code, Name: names[code]})
	}
	return languages
}

func hydrateRegistrationActivities(registrations []Registration) error {
	for i := range registrations {
		activities, err := hydrateActivitiesInOrder(registrations[i].Contest.allowedActivityIDs)
		if err != nil {
			return err
		}
		registrations[i].Contest.AllowedActivities = activities
	}
	return nil
}

func (s *Service) ValidateRegistrationUpsert(
	ctx context.Context,
	parameters RegistrationUpsertParameters,
) (*Contest, error) {
	contest, allowedLanguages, err := s.findContestWithLanguages(ctx, parameters.ContestID, false)
	if err != nil {
		return nil, err
	}

	if err := parameters.Validate(); err != nil {
		return nil, err
	}

	if len(allowedLanguages) > 0 {
		allowed := make(map[string]struct{}, len(allowedLanguages))
		for _, language := range allowedLanguages {
			allowed[language.Code] = struct{}{}
		}
		for _, code := range parameters.LanguageCodes {
			if _, ok := allowed[code]; !ok {
				return nil, errx.NewInvalidInputError("invalid contest registration LanguageCodes: contains a language not allowed by the contest")
			}
		}
	}

	return contest, nil
}

func (s *Service) ApplyRegistration(
	ctx context.Context,
	registration Registration,
	existing *Registration,
	contest Contest,
) error {
	removedLanguages := []string{}
	if existing != nil {
		selectedLanguages := make(map[string]struct{}, len(registration.LanguageCodes))
		for _, code := range registration.LanguageCodes {
			selectedLanguages[code] = struct{}{}
		}

		for _, language := range existing.Languages {
			if _, selected := selectedLanguages[language.Code]; !selected {
				removedLanguages = append(removedLanguages, language.Code)
			}
		}
	}

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

	return s.contests.FindContestByID(ctx, FindParameters{ID: contest.ID})
}

func (s *Service) CheckCreatePermission(ctx context.Context, userID uuid.UUID, accountCreatedAt time.Time) error {
	now := timex.Now()
	if err := checkContestCreatorAccountAge(accountCreatedAt, now); err != nil {
		return err
	}

	count, err := s.contests.CountContestsCreatedByUserForYear(ctx, userID, int32(now.Year()))
	if err != nil {
		return err
	}
	return checkContestCreationYearlyLimit(count)
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

// RequireExistingContest checks the stored contest and organizer without hydrating
// catalogs or introducing visibility rules into public log reads.
func (s *Service) RequireExistingContest(ctx context.Context, id uuid.UUID) error {
	_, err := s.contests.FindContestByID(ctx, FindParameters{ID: id})
	return err
}
