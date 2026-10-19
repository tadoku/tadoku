package identity

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type ReadService struct {
	repository *readRepository
	legacy     LegacyReader
	evaluator  *featureflags.Evaluator
}

var _ Reader = (*ReadService)(nil)

func NewReader(pool *pgxpool.Pool, legacy LegacyReader, evaluator *featureflags.Evaluator) *ReadService {
	return &ReadService{repository: &readRepository{pool: pool}, legacy: legacy, evaluator: evaluator}
}

func (s *ReadService) local(ctx context.Context) (bool, error) {
	if _, ok := tenant.FromContext(ctx); !ok {
		return false, errx.NewUnavailableError("identity tenant missing", nil)
	}
	return s.evaluator.TenantBoolean(ctx, featureflags.IdentityV2), nil
}

func (s *ReadService) FindAccount(ctx context.Context, userID uuid.UUID) (Account, error) {
	local, err := s.local(ctx)
	if err != nil {
		return Account{}, err
	}
	if local {
		return s.repository.find(ctx, userID)
	}
	return s.findLegacyAccount(ctx, userID)
}

func (s *ReadService) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	_, err := s.FindAccount(ctx, userID)
	if errx.KindOf(err) == errx.NotFound {
		return false, nil
	}
	return err == nil, err
}

type readCursor struct {
	Tenant string `json:"tenant"`
	Query  string `json:"query"`
	After  string `json:"after"`
}

func (s *ReadService) ListAccounts(ctx context.Context, query ListAccounts) (AccountPage, error) {
	local, err := s.local(ctx)
	if err != nil {
		return AccountPage{}, err
	}
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 500 {
		query.Limit = 500
	}
	if !local {
		return s.listLegacyAccounts(ctx, query)
	}

	key, _ := tenant.FromContext(ctx)
	after := uuid.Nil
	if query.Cursor != "" {
		raw, decodeErr := base64.RawURLEncoding.DecodeString(query.Cursor)
		var cursor readCursor
		if decodeErr != nil ||
			json.Unmarshal(raw, &cursor) != nil ||
			cursor.Tenant != key.String() ||
			cursor.Query != query.Query {
			return AccountPage{}, errx.NewInvalidInputError("invalid account cursor")
		}
		after, err = uuid.Parse(cursor.After)
		if err != nil {
			return AccountPage{}, errx.NewInvalidInputError("invalid account cursor")
		}
	}

	accounts, err := s.repository.list(ctx, query.Query, after, query.Limit+1)
	if err != nil {
		return AccountPage{}, err
	}
	page := AccountPage{Accounts: accounts}
	if len(accounts) > query.Limit {
		page.Accounts = accounts[:query.Limit]
		raw, _ := json.Marshal(readCursor{
			Tenant: key.String(),
			Query:  query.Query,
			After:  page.Accounts[query.Limit-1].ID.String(),
		})
		page.NextCursor = base64.RawURLEncoding.EncodeToString(raw)
	}
	return page, nil
}
