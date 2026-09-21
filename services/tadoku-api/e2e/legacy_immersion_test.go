package e2e_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	kratosapi "github.com/ory/kratos-client-go"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commonkratos "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/common/middleware"
	immersionory "github.com/tadoku/tadoku/services/immersion-api/client/ory"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/repository"
)

type legacyImmersionAPI struct {
	db                    *sql.DB
	handler               http.Handler
	scoringEnabledHandler http.Handler
}

func newLegacyImmersionAPI(ctx context.Context, dsn, jwksURL, ketoReadURL string, kratos *kratosapi.APIClient) (*legacyImmersionAPI, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDB(*config)
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}

	postgresRepository := repository.NewRepository(db)
	userUpsert := domain.NewUserUpsert(postgresRepository)

	kratosConfig := kratos.GetConfig()
	kratosClient := immersionory.NewKratosClient(
		kratosConfig.Servers[0].URL,
		commonkratos.WithHTTPClient(kratosConfig.HTTPClient),
	)

	handlers := make([]http.Handler, 0, 2)
	for _, enabled := range []bool{false, true} {
		server := rest.NewServer(
			domain.NewContestConfigurationOptions(postgresRepository),
			domain.NewLogConfigurationOptionsWithScoringEngine(postgresRepository, enabled),
			domain.NewContestFindLatestOfficial(postgresRepository),
			domain.NewContestSummaryFetch(postgresRepository),
			nil, // yearly activity split
			domain.NewContestFind(postgresRepository),
			nil, // log find
			domain.NewContestList(postgresRepository),
			nil, // user logs
			nil, // contest logs
			domain.NewRegistrationFind(postgresRepository),
			domain.NewRegistrationListYearly(postgresRepository),
			nil, // contest leaderboard
			nil, // yearly leaderboard
			nil, // global leaderboard
			nil, // profile contest
			nil, // profile contest activity
			nil, // profile yearly activity
			nil, // profile yearly scores
			nil, // profile fetch
			domain.NewRegistrationListOngoing(postgresRepository, scenarioClock{}),
			domain.NewContestPermissionCheck(postgresRepository, kratosClient, scenarioClock{}),
			nil, // log delete
			nil, // moderation detach log
			domain.NewRegistrationUpsert(postgresRepository, userUpsert),
			nil, // log create
			nil, // log update
			domain.NewContestCreate(postgresRepository, scenarioClock{}, userUpsert),
			domain.NewLanguageList(postgresRepository),
			domain.NewLanguageCreate(postgresRepository),
			domain.NewLanguageUpdate(postgresRepository),
			domain.NewTagSuggestions(postgresRepository),
			nil, // log contest update
			nil, // score preview
			nil, // scoring rule set management
			nil, // feature flags
			nil, // feature access
		)
		router := echo.New()
		router.Logger.SetOutput(io.Discard)
		router.Use(echomiddleware.Recover())

		roles := commonroles.NewKetoService(ketoclient.NewReadClient(ketoReadURL), "app", "tadoku")
		api := router.Group("/immersion",
			middleware.VerifyJWT(jwksURL),
			middleware.Identity(),
			middleware.RolesFromKeto(roles),
			middleware.RequireServiceAudience("immersion-api"),
			middleware.RejectBannedUsers(),
		)

		openapi.RegisterHandlers(api, server)

		handlers = append(handlers, router)
	}
	return &legacyImmersionAPI{db: db, handler: handlers[0], scoringEnabledHandler: handlers[1]}, nil
}
