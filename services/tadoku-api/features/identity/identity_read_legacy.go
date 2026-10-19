package identity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	kratosapi "github.com/ory/kratos-client-go"
	"github.com/sahilm/fuzzy"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func legacyAccount(value kratosapi.Identity) (Account, error) {
	if value.GetSchemaId() != "user" {
		return Account{}, errx.NewNotFoundError("identity not found")
	}
	var traits struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	}
	raw, err := json.Marshal(value.GetTraits())
	if err != nil {
		return Account{}, err
	}
	if err := json.Unmarshal(raw, &traits); err != nil {
		return Account{}, err
	}
	id, err := uuid.Parse(value.GetId())
	if err != nil {
		return Account{}, err
	}
	return Account{
		ID:          id,
		Email:       Email{value: traits.Email},
		DisplayName: DisplayName{value: traits.DisplayName},
		CreatedAt:   value.GetCreatedAt(),
	}, nil
}

func (s *ReadService) findLegacyAccount(ctx context.Context, userID uuid.UUID) (Account, error) {
	value, err := s.legacy.FetchIdentity(ctx, userID)
	if errors.Is(err, ErrLegacyNotFound) {
		return Account{}, errx.NewNotFoundError("identity not found")
	}
	if err != nil {
		return Account{}, errx.NewUnavailableError("fetch identity", err)
	}
	return legacyAccount(*value)
}

type legacySearch []Account

func (s legacySearch) Len() int { return len(s) }
func (s legacySearch) String(i int) string {
	return strings.ToLower(s[i].DisplayName.String() + " " + s[i].Email.String())
}

func (s *ReadService) listLegacyAccounts(ctx context.Context, query ListAccounts) (AccountPage, error) {
	accounts := make([]Account, 0)
	seenIDs := make(map[uuid.UUID]bool)
	seenTokens := make(map[string]bool)
	token := ""
	for {
		values, next, err := s.legacy.ListIdentities(ctx, 500, token)
		if err != nil {
			return AccountPage{}, errx.NewUnavailableError("list users", err)
		}
		for _, value := range values {
			account, err := legacyAccount(value)
			if err != nil || seenIDs[account.ID] {
				continue
			}
			seenIDs[account.ID] = true
			accounts = append(accounts, account)
		}
		if next == "" {
			break
		}
		if seenTokens[next] {
			return AccountPage{}, errx.NewUnavailableError("list users", fmt.Errorf("repeated next page token"))
		}
		seenTokens[next] = true
		token = next
	}
	if query.Query != "" {
		matches := fuzzy.FindFrom(strings.ToLower(query.Query), legacySearch(accounts))
		filtered := make([]Account, 0, len(matches))
		for _, match := range matches {
			filtered = append(filtered, accounts[match.Index])
		}
		accounts = filtered
	}
	offset := 0
	if query.Cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(query.Cursor)
		if err != nil {
			return AccountPage{}, errx.NewInvalidInputError("invalid account cursor")
		}
		offset, err = strconv.Atoi(string(raw))
		if err != nil || offset < 0 {
			return AccountPage{}, errx.NewInvalidInputError("invalid account cursor")
		}
	}
	if offset >= len(accounts) {
		return AccountPage{Accounts: []Account{}}, nil
	}
	end := offset + min(query.Limit, len(accounts)-offset)
	page := AccountPage{Accounts: accounts[offset:end]}
	if end < len(accounts) {
		page.NextCursor = base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(end)))
	}
	return page, nil
}
