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
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/content-api/domain"
	"github.com/tadoku/tadoku/services/content-api/http/rest"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/content-api/storage/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type legacyContentAPI struct {
	db      *sql.DB
	handler http.Handler
}

func newLegacyContentAPI(ctx context.Context, dsn, jwksURL, ketoReadURL string) (*legacyContentAPI, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDB(*config)
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}

	pageRepository := postgres.NewPageRepository(db)
	repository := postgres.NewAnnouncementRepository(db)
	postsRepository := postgres.NewPostRepository(db)
	server := rest.NewServer(
		domain.NewPageCreate(pageRepository, scenarioClock{}),
		domain.NewPageUpdate(pageRepository, scenarioClock{}),
		domain.NewPageDelete(pageRepository),
		domain.NewPageFind(pageRepository, scenarioClock{}),
		domain.NewPageFindByID(pageRepository),
		domain.NewPageList(pageRepository),
		domain.NewPageVersionList(pageRepository),
		domain.NewPageVersionGet(pageRepository),
		domain.NewPostCreate(postsRepository, scenarioClock{}),
		domain.NewPostUpdate(postsRepository, scenarioClock{}),
		domain.NewPostDelete(postsRepository),
		domain.NewPostFind(postsRepository, scenarioClock{}),
		domain.NewPostFindByID(postsRepository),
		domain.NewPostList(postsRepository),
		domain.NewPostVersionList(postsRepository),
		domain.NewPostVersionGet(postsRepository),
		domain.NewAnnouncementCreate(repository, scenarioClock{}),
		domain.NewAnnouncementUpdate(repository, scenarioClock{}),
		domain.NewAnnouncementDelete(repository),
		domain.NewAnnouncementFindByID(repository),
		domain.NewAnnouncementList(repository),
		domain.NewAnnouncementListActive(repository, scenarioClock{}),
	)
	router := echo.New()
	router.Logger.SetOutput(io.Discard)
	roles := commonroles.NewKetoService(ketoclient.NewReadClient(ketoReadURL), "app", "tadoku")
	// Reuse production route registration and its business authentication stack.
	api := router.Group("/content",
		middleware.VerifyJWT(jwksURL),
		middleware.Identity(),
		middleware.RolesFromKeto(roles),
		middleware.RequireServiceAudience("content-api"),
		middleware.RejectBannedUsers(),
	)
	openapi.RegisterHandlers(api, server)

	return &legacyContentAPI{db: db, handler: router}, nil
}

type scenarioClock struct{}

func (scenarioClock) Now() time.Time { return timex.Now() }
