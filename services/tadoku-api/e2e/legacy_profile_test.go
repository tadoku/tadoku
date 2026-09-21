package e2e_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	kratosclient "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/common/middleware"
	legacycache "github.com/tadoku/tadoku/services/profile-api/cache"
	legacyory "github.com/tadoku/tadoku/services/profile-api/client/ory"
	legacydomain "github.com/tadoku/tadoku/services/profile-api/domain"
	legacyrest "github.com/tadoku/tadoku/services/profile-api/http/rest"
	legacyopenapi "github.com/tadoku/tadoku/services/profile-api/http/rest/openapi"
	legacyrepository "github.com/tadoku/tadoku/services/profile-api/storage/postgres/repository"
)

type legacyProfileAPI struct {
	db         *sql.DB
	handler    http.Handler
	cache      *legacyProfileCacheHolder
	kratos     *legacyory.KratosClient
	repository *legacyrepository.Repository
}

type legacyProfileCacheHolder struct {
	mu    sync.Mutex
	cache *legacycache.UserCache
	warm  bool
}

func (h *legacyProfileCacheHolder) GetUsers() []legacydomain.UserCacheEntry {
	h.mu.Lock()
	cache := h.cache
	h.mu.Unlock()
	if cache == nil {
		return []legacydomain.UserCacheEntry{}
	}
	return cache.GetUsers()
}

func (h *legacyProfileCacheHolder) replace(cache *legacycache.UserCache) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cache = cache
	h.warm = false
}

func (h *legacyProfileCacheHolder) warmCache(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.warm {
		return nil
	}

	if err := h.cache.Refresh(ctx); err != nil {
		return err
	}
	h.warm = true
	return nil
}

func newLegacyProfileAPI(ctx context.Context, dsn, jwksURL, ketoReadURL string, cursorClient *kratosclient.Client) (*legacyProfileAPI, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDB(*config)
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}

	roles := commonroles.NewKetoService(ketoclient.NewReadClient(ketoReadURL), "app", "tadoku")
	holder := &legacyProfileCacheHolder{}
	server := legacyrest.NewServer(legacydomain.NewUserList(holder, roles))
	router := echo.New()
	middleware.RestoreJSONCharset(router)
	router.Logger.SetOutput(io.Discard)
	api := router.Group("/profile",
		middleware.VerifyJWT(jwksURL),
		middleware.Identity(),
		middleware.RolesFromKeto(roles),
		middleware.RequireServiceAudience("profile-api"),
		middleware.RejectBannedUsers(),
	)
	legacyopenapi.RegisterHandlers(api, server)

	repository := legacyrepository.NewRepository(db)
	legacy := &legacyProfileAPI{
		db:         db,
		cache:      holder,
		kratos:     legacyory.NewKratosClientFromClient(cursorClient),
		repository: repository,
	}
	legacy.resetCache()
	legacy.handler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if err := holder.warmCache(request.Context()); err != nil {
			slog.ErrorContext(request.Context(), "refresh legacy profile cache", "error", err)
		}
		router.ServeHTTP(response, request)
	})
	return legacy, nil
}

func (a *legacyProfileAPI) resetCache() {
	a.cache.replace(legacycache.NewUserCache(a.kratos, a.repository, time.Hour))
}
