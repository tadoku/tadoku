package e2e_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/authz-api/domain"
	"github.com/tadoku/tadoku/services/authz-api/http/rest"
	legacyopenapi "github.com/tadoku/tadoku/services/authz-api/http/rest/openapi"
	legacyrepository "github.com/tadoku/tadoku/services/authz-api/storage/postgres/repository"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	kratosclient "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/common/middleware"
)

type legacyAuthzAPI struct {
	handler http.Handler
	db      *sql.DB
}

func newLegacyAuthzAPI(
	ctx context.Context,
	dsn string,
	jwksURL string,
	ketoReadURL string,
	ketoWriteURL string,
	kratos *kratosclient.Client,
) (*legacyAuthzAPI, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDB(*config)
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	handler, err := newLegacyAuthzHandler(
		jwksURL,
		ketoReadURL,
		ketoWriteURL,
		kratos,
		legacyrepository.NewRepository(db),
	)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return &legacyAuthzAPI{handler: handler, db: db}, nil
}

func newLegacyAuthzHandler(
	jwksURL string,
	ketoReadURL string,
	ketoWriteURL string,
	kratos *kratosclient.Client,
	audit domain.ModerationAuditRepository,
) (http.Handler, error) {
	keto := ketoclient.NewClient(ketoReadURL, ketoWriteURL)
	roles := commonroles.NewKetoService(keto, "app", "tadoku")
	allowlist, err := domain.ParsePermissionAllowlist("")
	if err != nil {
		return nil, fmt.Errorf("parse public permission allowlist: %w", err)
	}

	server := rest.NewServer(
		domain.NewRoleGet(roles),
		domain.NewRoleUpdate(kratos, audit, roles, commonroles.NewKetoManager(keto, "app", "tadoku")),
		domain.NewPublicPermissionCheck(keto, allowlist),
		nil,
		nil,
		nil,
	)
	router := echo.New()
	router.Logger.SetOutput(io.Discard)
	api := router.Group("/authz",
		middleware.VerifyJWT(jwksURL),
		middleware.Identity(),
		middleware.RolesFromKeto(roles),
		middleware.RequireServiceAudience("authz-api"),
	)
	legacyopenapi.RegisterHandlers(api, server)

	return router, nil
}
