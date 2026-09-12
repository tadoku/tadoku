// Package legacy is a test-only compatibility fixture. Native runtime code must
// never depend on this package, Echo, or the legacy generated server bindings.
package legacy

import (
	"database/sql"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/tadoku/tadoku/services/common/authz/roles"
	"github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/content-api/domain"
	"github.com/tadoku/tadoku/services/content-api/http/rest"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/content-api/storage/postgres"
)

func ContentHandler(db *sql.DB, jwksURL, ketoURL string) http.Handler {
	server := rest.NewServer(nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		domain.NewAnnouncementListActive(postgres.NewAnnouncementRepository(db)))
	e := echo.New()
	e.Logger.SetOutput(io.Discard)
	e.Use(echomiddleware.Recover(), middleware.VerifyJWT(jwksURL), middleware.Identity(),
		middleware.RolesFromKeto(roles.NewKetoService(keto.NewReadClient(ketoURL), "app", "tadoku")),
		middleware.RequireServiceAudience("content-api"), middleware.RejectBannedUsers())
	openapi.RegisterHandlers(e, server)
	return e
}
