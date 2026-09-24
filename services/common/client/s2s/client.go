package s2s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type Client struct {
	oathkeeperURL string
	k8sTokenPath  string
	httpClient    *http.Client
	now           func() time.Time

	mu         sync.RWMutex
	tokenCache map[string]*cachedToken
}

type cachedToken struct {
	token     string
	expiresAt time.Time
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		if httpClient != nil {
			client.httpClient = httpClient
		}
	}
}

func WithTokenPath(path string) Option {
	return func(client *Client) {
		if path != "" {
			client.k8sTokenPath = path
		}
	}
}

func NewClient(oathkeeperURL string, options ...Option) *Client {
	client := &Client{
		oathkeeperURL: oathkeeperURL,
		k8sTokenPath:  "/var/run/secrets/tokens/token",
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		now:           time.Now,
		tokenCache:    make(map[string]*cachedToken),
	}
	for _, option := range options {
		option(client)
	}
	return client
}

func (c *Client) GetToken(targetService string) (string, error) {
	return c.GetTokenContext(context.Background(), targetService)
}

func (c *Client) GetTokenContext(ctx context.Context, targetService string) (string, error) {
	c.mu.RLock()
	if cached, ok := c.tokenCache[targetService]; ok {
		if c.now().Before(cached.expiresAt) {
			c.mu.RUnlock()
			return cached.token, nil
		}
	}
	c.mu.RUnlock()

	k8sToken, err := os.ReadFile(c.k8sTokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read k8s token: %w", err)
	}

	url := fmt.Sprintf("%s/token-exchange/%s", c.oathkeeperURL, targetService)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	trimmedToken := strings.TrimSpace(string(k8sToken))
	req.Header.Set("Authorization", "Bearer "+trimmedToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" || !strings.EqualFold(tokenResp.TokenType, "bearer") || tokenResp.ExpiresIn <= 0 {
		return "", fmt.Errorf("token exchange returned an invalid bearer token")
	}

	expiresIn := tokenResp.ExpiresIn
	cacheSeconds := expiresIn - 300
	if cacheSeconds <= 0 {
		cacheSeconds = expiresIn
	}

	c.mu.Lock()
	c.tokenCache[targetService] = &cachedToken{
		token:     tokenResp.AccessToken,
		expiresAt: c.now().Add(time.Duration(cacheSeconds) * time.Second),
	}
	c.mu.Unlock()

	return tokenResp.AccessToken, nil
}
