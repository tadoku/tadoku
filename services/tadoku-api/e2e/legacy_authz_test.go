package e2e_test

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/authz-api/domain"
	"github.com/tadoku/tadoku/services/authz-api/http/rest"
	legacyopenapi "github.com/tadoku/tadoku/services/authz-api/http/rest/openapi"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/common/middleware"
)

type legacyAuthzAPI struct {
	handler http.Handler
}

func newLegacyAuthzAPI(jwksURL, ketoReadURL string) (*legacyAuthzAPI, error) {
	keto := ketoclient.NewReadClient(ketoReadURL)
	roles := commonroles.NewKetoService(keto, "app", "tadoku")
	allowlist, err := domain.ParsePermissionAllowlist("")
	if err != nil {
		return nil, fmt.Errorf("parse public permission allowlist: %w", err)
	}

	server := rest.NewServer(
		domain.NewRoleGet(roles),
		nil,
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

	return &legacyAuthzAPI{handler: router}, nil
}
