package s2s

import (
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

	mu         sync.RWMutex
	tokenCache map[string]*cachedToken
}

type cachedToken struct {
	token     string
	expiresAt time.Time
}

func NewClient(oathkeeperURL string) *Client {
	return &Client{
		oathkeeperURL: oathkeeperURL,
		k8sTokenPath:  "/var/run/secrets/tokens/token",
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		tokenCache:    make(map[string]*cachedToken),
	}
}

// GetToken returns a JWT for calling the target service.
func (c *Client) GetToken(targetService string) (string, error) {
	c.mu.RLock()
	if cached, ok := c.tokenCache[targetService]; ok {
		if time.Now().Before(cached.expiresAt) {
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
	req, err := http.NewRequest("GET", url, nil)
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

	expiresIn := tokenResp.ExpiresIn
	cacheSeconds := expiresIn - 300
	if cacheSeconds <= 0 {
		cacheSeconds = expiresIn
	}

	c.mu.Lock()
	c.tokenCache[targetService] = &cachedToken{
		token:     tokenResp.AccessToken,
		expiresAt: time.Now().Add(time.Duration(cacheSeconds) * time.Second),
	}
	c.mu.Unlock()

	return tokenResp.AccessToken, nil
}
