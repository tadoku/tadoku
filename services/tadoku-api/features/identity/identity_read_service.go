package identity

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type ReadService struct {
	repository *readRepository
}

var _ Reader = (*ReadService)(nil)

func NewReader(pool *pgxpool.Pool) *ReadService {
	return &ReadService{repository: &readRepository{pool: pool}}
}

func (s *ReadService) FindAccount(ctx context.Context, userID uuid.UUID) (Account, error) {
	return s.repository.find(ctx, userID)
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
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 500 {
		query.Limit = 500
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
		parsed, err := uuid.Parse(cursor.After)
		if err != nil {
			return AccountPage{}, errx.NewInvalidInputError("invalid account cursor")
		}
		after = parsed
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

func (s *ReadService) ListAllAccounts(ctx context.Context, query string) ([]Account, error) {
	accounts := make([]Account, 0)
	cursor := ""
	for {
		batch, err := s.ListAccounts(ctx, ListAccounts{Query: query, Limit: 500, Cursor: cursor})
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, batch.Accounts...)
		if batch.NextCursor == "" {
			return accounts, nil
		}
		cursor = batch.NextCursor
	}
}
