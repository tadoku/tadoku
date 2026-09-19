package kratos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	kratosapi "github.com/ory/kratos-client-go"
)

var ErrNotFound = errors.New("kratos identity not found")

type Client struct {
	client            *kratosapi.APIClient
	httpClient        *http.Client
	listIdentitiesURL string
}

func NewClient(kratosURL string) *Client {
	cfg := kratosapi.NewConfiguration()
	cfg.Servers = kratosapi.ServerConfigurations{{URL: kratosURL}}
	return &Client{
		client:            kratosapi.NewAPIClient(cfg),
		httpClient:        http.DefaultClient,
		listIdentitiesURL: strings.TrimRight(kratosURL, "/") + "/admin/identities",
	}
}

func (c *Client) FetchIdentity(ctx context.Context, id uuid.UUID) (*kratosapi.Identity, error) {
	req := c.client.IdentityApi.GetIdentity(ctx, id.String())
	identity, res, err := c.client.IdentityApi.GetIdentityExecute(req)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch identity: %w", err)
	}
	return identity, nil
}

func (c *Client) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := c.FetchIdentity(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Client) ListIdentities(ctx context.Context) ([]kratosapi.Identity, error) {
	const pageSize = 500

	var identities []kratosapi.Identity
	pageToken := ""
	seenPageTokens := make(map[string]struct{})

	for {
		requestURL, err := url.Parse(c.listIdentitiesURL)
		if err != nil {
			return nil, fmt.Errorf("could not list identities: %w", err)
		}
		query := requestURL.Query()
		query.Set("page_size", fmt.Sprint(pageSize))
		if pageToken != "" {
			query.Set("page_token", pageToken)
		}
		requestURL.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("could not list identities: %w", err)
		}
		req.Header.Set("Accept", "application/json")

		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not list identities: %w", err)
		}
		if res.StatusCode >= http.StatusMultipleChoices {
			_ = res.Body.Close()
			return nil, fmt.Errorf("could not list identities: %s", res.Status)
		}

		var page []kratosapi.Identity
		decodeErr := json.NewDecoder(res.Body).Decode(&page)
		closeErr := res.Body.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("could not list identities: %w", decodeErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("could not list identities: %w", closeErr)
		}
		identities = append(identities, page...)

		nextPageToken, hasNext, err := nextPageToken(res)
		if err != nil {
			return nil, fmt.Errorf("could not list identities: %w", err)
		}
		if !hasNext {
			return identities, nil
		}
		if _, seen := seenPageTokens[nextPageToken]; seen {
			return nil, fmt.Errorf("could not list identities: repeated next page token")
		}
		seenPageTokens[nextPageToken] = struct{}{}
		pageToken = nextPageToken
	}
}

func nextPageToken(res *http.Response) (string, bool, error) {
	for _, header := range res.Header.Values("Link") {
		for _, link := range strings.Split(header, ",") {
			parts := strings.Split(link, ";")
			if len(parts) < 2 || !hasNextRelation(parts[1:]) {
				continue
			}

			target := strings.TrimSpace(parts[0])
			if len(target) < 2 || target[0] != '<' || target[len(target)-1] != '>' {
				return "", false, fmt.Errorf("invalid next page link")
			}
			parsed, err := url.Parse(target[1 : len(target)-1])
			if err != nil {
				return "", false, fmt.Errorf("invalid next page link: %w", err)
			}
			pageTokens, ok := parsed.Query()["page_token"]
			if !ok || len(pageTokens) != 1 || pageTokens[0] == "" {
				return "", false, fmt.Errorf("next page link has no page_token")
			}
			return pageTokens[0], true, nil
		}
	}
	return "", false, nil
}

func hasNextRelation(parameters []string) bool {
	for _, parameter := range parameters {
		name, value, ok := strings.Cut(strings.TrimSpace(parameter), "=")
		if !ok || !strings.EqualFold(name, "rel") {
			continue
		}
		for _, relation := range strings.Fields(strings.Trim(value, `"`)) {
			if relation == "next" {
				return true
			}
		}
	}
	return false
}

func (c *Client) DeactivateIdentity(ctx context.Context, id uuid.UUID) error {
	statePatch := kratosapi.NewJsonPatch("replace", "/state")
	statePatch.SetValue(kratosapi.IDENTITYSTATE_INACTIVE)
	req := c.client.IdentityApi.PatchIdentity(ctx, id.String()).JsonPatch([]kratosapi.JsonPatch{*statePatch})
	_, res, err := c.client.IdentityApi.PatchIdentityExecute(req)
	if err != nil {
		if responseStatus(res, http.StatusNotFound) {
			return nil
		}
		return fmt.Errorf("could not deactivate identity: %w", err)
	}
	return nil
}

func (c *Client) DeleteIdentitySessions(ctx context.Context, id uuid.UUID) error {
	req := c.client.IdentityApi.DeleteIdentitySessions(ctx, id.String())
	res, err := c.client.IdentityApi.DeleteIdentitySessionsExecute(req)
	if err != nil {
		if responseStatus(res, http.StatusNotFound) {
			return nil
		}
		return fmt.Errorf("could not delete identity sessions: %w", err)
	}
	return nil
}

func (c *Client) DeleteIdentity(ctx context.Context, id uuid.UUID) error {
	req := c.client.IdentityApi.DeleteIdentity(ctx, id.String())
	res, err := c.client.IdentityApi.DeleteIdentityExecute(req)
	if err != nil {
		if responseStatus(res, http.StatusNotFound) {
			return nil
		}
		return fmt.Errorf("could not delete identity: %w", err)
	}
	return nil
}

func responseStatus(res *http.Response, status int) bool {
	return res != nil && res.StatusCode == status
}
