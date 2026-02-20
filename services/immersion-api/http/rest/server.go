package rest

import (
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	contestConfigurationOptions *domain.ContestConfigurationOptions,
	logConfigurationOptions *domain.LogConfigurationOptions,
	contestFindLatestOfficial *domain.ContestFindLatestOfficial,
	contestSummaryFetch *domain.ContestSummaryFetch,
	profileYearlyActivitySplit *domain.ProfileYearlyActivitySplit,
	contestFind *domain.ContestFind,
	logFind *domain.LogFind,
	contestList *domain.ContestList,
	logListForUser *domain.LogListForUser,
	logListForContest *domain.LogListForContest,
	registrationFind *domain.RegistrationFind,
	registrationListYearly *domain.RegistrationListYearly,
	contestLeaderboardFetch *domain.ContestLeaderboardFetch,
	leaderboardYearly *domain.LeaderboardYearly,
	leaderboardGlobal *domain.LeaderboardGlobal,
	profileContest *domain.ProfileContest,
	profileContestActivity *domain.ProfileContestActivity,
	profileYearlyActivity *domain.ProfileYearlyActivity,
	profileYearlyScores *domain.ProfileYearlyScores,
	profileFetch *domain.ProfileFetch,
	registrationListOngoing *domain.RegistrationListOngoing,
	contestPermissionCheck *domain.ContestPermissionCheck,
	logDelete *domain.LogDelete,
	contestModerationDetachLog *domain.ContestModerationDetachLog,
	registrationUpsert *domain.RegistrationUpsert,
	logCreate *domain.LogCreate,
	logUpdate *domain.LogUpdate,
	contestCreate *domain.ContestCreate,
	languageList *domain.LanguageList,
	languageCreate *domain.LanguageCreate,
	languageUpdate *domain.LanguageUpdate,
	tagSuggestions *domain.TagSuggestions,
	logContestUpdate *domain.LogContestUpdate,
) openapi.ServerInterface {
	return &Server{
		contestConfigurationOptions: contestConfigurationOptions,
		logConfigurationOptions:     logConfigurationOptions,
		contestFindLatestOfficial:   contestFindLatestOfficial,
		contestSummaryFetch:         contestSummaryFetch,
		profileYearlyActivitySplit:  profileYearlyActivitySplit,
		contestFind:                 contestFind,
		logFind:                     logFind,
		contestList:                 contestList,
		logListForUser:              logListForUser,
		logListForContest:           logListForContest,
		registrationFind:            registrationFind,
		registrationListYearly:      registrationListYearly,
		contestLeaderboardFetch:     contestLeaderboardFetch,
		leaderboardYearly:           leaderboardYearly,
		leaderboardGlobal:           leaderboardGlobal,
		profileContest:              profileContest,
		profileContestActivity:      profileContestActivity,
		profileYearlyActivity:       profileYearlyActivity,
		profileYearlyScores:         profileYearlyScores,
		profileFetch:                profileFetch,
		registrationListOngoing:     registrationListOngoing,
		contestPermissionCheck:      contestPermissionCheck,
		logDelete:                   logDelete,
		contestModerationDetachLog:  contestModerationDetachLog,
		registrationUpsert:          registrationUpsert,
		logCreate:                   logCreate,
		logUpdate:                   logUpdate,
		contestCreate:               contestCreate,
		languageList:                languageList,
		languageCreate:              languageCreate,
		languageUpdate:              languageUpdate,
		tagSuggestions:              tagSuggestions,
		logContestUpdate:            logContestUpdate,
	}
}

type Server struct {
	contestConfigurationOptions *domain.ContestConfigurationOptions
	logConfigurationOptions     *domain.LogConfigurationOptions
	contestFindLatestOfficial   *domain.ContestFindLatestOfficial
	contestSummaryFetch         *domain.ContestSummaryFetch
	profileYearlyActivitySplit  *domain.ProfileYearlyActivitySplit
	contestFind                 *domain.ContestFind
	logFind                     *domain.LogFind
	contestList                 *domain.ContestList
	logListForUser              *domain.LogListForUser
	logListForContest           *domain.LogListForContest
	registrationFind            *domain.RegistrationFind
	registrationListYearly      *domain.RegistrationListYearly
	contestLeaderboardFetch     *domain.ContestLeaderboardFetch
	leaderboardYearly           *domain.LeaderboardYearly
	leaderboardGlobal           *domain.LeaderboardGlobal
	profileContest              *domain.ProfileContest
	profileContestActivity      *domain.ProfileContestActivity
	profileYearlyActivity       *domain.ProfileYearlyActivity
	profileYearlyScores         *domain.ProfileYearlyScores
	profileFetch                *domain.ProfileFetch
	registrationListOngoing     *domain.RegistrationListOngoing
	contestPermissionCheck      *domain.ContestPermissionCheck
	logDelete                   *domain.LogDelete
	contestModerationDetachLog  *domain.ContestModerationDetachLog
	registrationUpsert          *domain.RegistrationUpsert
	logCreate                   *domain.LogCreate
	logUpdate                   *domain.LogUpdate
	contestCreate               *domain.ContestCreate
	languageList                *domain.LanguageList
	languageCreate              *domain.LanguageCreate
	languageUpdate              *domain.LanguageUpdate
	tagSuggestions              *domain.TagSuggestions
	logContestUpdate            *domain.LogContestUpdate
}
