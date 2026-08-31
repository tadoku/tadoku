package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type JWKSProxy struct {
	client *http.Client
	token  string
}

var errExpiredToken = errors.New("token is expired")

func main() {
	jwksProxy := newJWKSProxy()
	clock, err := commondomain.NewClock("UTC")
	if err != nil {
		panic(err)
	}

	http.Handle("/", newTokenHandler(clock))

	http.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		body, err := jwksProxy.Fetch(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	_ = http.ListenAndServe(":"+port, nil)
}

func newTokenHandler(clock commondomain.Clock) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Id-Token")
		if token == "" {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				token = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if token == "" {
			http.Error(w, "missing X-Id-Token header", http.StatusBadRequest)
			return
		}

		expiresIn, err := tokenExpiresIn(token, clock.Now())
		if err != nil {
			if errors.Is(err, errExpiredToken) {
				http.Error(w, "token is expired", http.StatusUnauthorized)
				return
			}
			http.Error(w, "token has an invalid expiry", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
		})
	})
}

func tokenExpiresIn(token string, now time.Time) (int, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("token must have three segments")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, fmt.Errorf("decode token payload: %w", err)
	}

	var claims struct {
		ExpiresAt json.Number `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, fmt.Errorf("decode token claims: %w", err)
	}
	if claims.ExpiresAt == "" {
		return 0, fmt.Errorf("token is missing exp claim")
	}

	expiresAt, err := claims.ExpiresAt.Int64()
	if err != nil {
		return 0, fmt.Errorf("decode exp claim: %w", err)
	}

	remaining := time.Unix(expiresAt, 0).Sub(now)
	if remaining <= 0 {
		return 0, errExpiredToken
	}

	return int(remaining / time.Second), nil
}

func newJWKSProxy() *JWKSProxy {
	token := ""
	tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err == nil {
		token = strings.TrimSpace(string(tokenBytes))
	}

	pool := x509.NewCertPool()
	if caBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"); err == nil {
		pool.AppendCertsFromPEM(caBytes)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}

	return &JWKSProxy{
		client: client,
		token:  token,
	}
}

func (p *JWKSProxy) Fetch(ctx context.Context) ([]byte, error) {
	if p.token == "" {
		return nil, fmt.Errorf("missing service account token")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://kubernetes.default.svc/openid/v1/jwks", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jwks fetch failed: %s - %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return io.ReadAll(resp.Body)
}
