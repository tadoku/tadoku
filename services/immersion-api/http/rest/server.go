package rest

import (
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	commandService command.Service,
	queryService query.Service,
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
) openapi.ServerInterface {
	return &Server{
		commandService:              commandService,
		queryService:                queryService,
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
	}
}

type Server struct {
	commandService command.Service
	queryService   query.Service

	// Service-per-function services
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
}
