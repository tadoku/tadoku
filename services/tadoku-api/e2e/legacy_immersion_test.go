package e2e_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	kratosapi "github.com/ory/kratos-client-go"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commonkratos "github.com/tadoku/tadoku/services/common/client/kratos"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/featureflags"
	"github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/immersion-api/client/fliptmanagement"
	immersionory "github.com/tadoku/tadoku/services/immersion-api/client/ory"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/repository"
	valkeystore "github.com/tadoku/tadoku/services/immersion-api/storage/valkey"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	valkeygo "github.com/valkey-io/valkey-go"
)

type scenarioClock struct{}

func (scenarioClock) Now() time.Time { return timex.Now() }

type legacyImmersionAPI struct {
	db                    *sql.DB
	handler               http.Handler
	scoringEnabledHandler http.Handler
}

func newLegacyImmersionAPI(ctx context.Context, dsn, jwksURL, ketoReadURL string, kratos *kratosapi.APIClient, valkeyClient valkeygo.Client) (*legacyImmersionAPI, error) {
	return newLegacyImmersionAPIWithTimeout(ctx, dsn, jwksURL, ketoReadURL, kratos, valkeyClient, time.Second)
}

func newLegacyImmersionAPIWithTimeout(ctx context.Context, dsn, jwksURL, ketoReadURL string, kratos *kratosapi.APIClient, valkeyClient valkeygo.Client, valkeyTimeout time.Duration) (*legacyImmersionAPI, error) {
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
	featureAccess := domain.NewFeatureAccess(fliptmanagement.NewClient(fliptmanagement.Config{URL: flipt.URL(), Environment: "local"}), postgresRepository)
	featureFlagEvaluator := featureflags.NewEvaluator(flipt, nil, commondomain.NewMockClock(fixtureInstant))
	leaderboardStore := valkeystore.NewLeaderboardStore(valkeyClient, scenarioClock{}, valkeyTimeout)

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
			domain.NewProfileYearlyActivitySplit(postgresRepository),
			domain.NewContestFind(postgresRepository),
			domain.NewLogFind(postgresRepository),
			domain.NewContestList(postgresRepository),
			domain.NewLogListForUser(postgresRepository),
			domain.NewLogListForContest(postgresRepository),
			domain.NewRegistrationFind(postgresRepository),
			domain.NewRegistrationListYearly(postgresRepository),
			domain.NewContestLeaderboardFetch(postgresRepository, leaderboardStore),
			domain.NewLeaderboardYearly(postgresRepository, leaderboardStore),
			domain.NewLeaderboardGlobal(postgresRepository, leaderboardStore),
			domain.NewProfileContest(postgresRepository),
			domain.NewProfileContestActivity(postgresRepository),
			domain.NewProfileYearlyActivity(postgresRepository),
			domain.NewProfileYearlyScores(postgresRepository),
			domain.NewProfileFetch(kratosClient),
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
			featureFlagEvaluator,
			featureAccess,
		)
		router := echo.New()
		middleware.RestoreJSONCharset(router)
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
