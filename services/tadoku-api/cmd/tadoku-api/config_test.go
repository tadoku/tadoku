package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadConfigUsesValidatedDefaults(t *testing.T) {
	t.Setenv("API_SCORING_ENGINE_ENABLED", "false")
	t.Setenv("API_JWKS", "http://jwks.test")
	t.Setenv("API_OATHKEEPER_AUTHZ_TOKEN", "callback-token")
	t.Setenv("API_KRATOS_ADMIN_URL", "http://kratos-admin.test")
	t.Setenv("API_IMMERSION_URL", "")
	t.Setenv("API_VALKEY_URL", "redis://valkey:6379")
	for key, value := range map[string]string{"HOST": "localhost", "DATABASE": "tadoku", "USER": "tadoku", "PASSWORD": "synthetic", "SSLMODE": "disable"} {
		t.Setenv("API_POSTGRES_"+key, value)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8000 {
		t.Errorf("got %v, want %v", cfg.Port, 8000)
	}
	if cfg.MetricsPort != 9090 {
		t.Errorf("got %v, want %v", cfg.MetricsPort, 9090)
	}
	if cfg.ServiceName != "tadoku-api" {
		t.Errorf("got %v, want %v", cfg.ServiceName, "tadoku-api")
	}
	if cfg.RequestTimeout != 30*time.Second {
		t.Errorf("got %v, want %v", cfg.RequestTimeout, 30*time.Second)
	}
	if cfg.MaxTokenAge != 24*time.Hour {
		t.Errorf("maximum token age=%v, want %v", cfg.MaxTokenAge, 24*time.Hour)
	}
	if cfg.JWTIssuer != "" {
		t.Errorf("JWT issuer=%q, want empty", cfg.JWTIssuer)
	}
	if cfg.PostgresMaxConnections != 4 {
		t.Errorf("pool limit=%d want=4", cfg.PostgresMaxConnections)
	}
	if cfg.ValkeyURL != "redis://valkey:6379" || cfg.ValkeyTimeout != time.Second {
		t.Errorf("Valkey URL=%q timeout=%v", cfg.ValkeyURL, cfg.ValkeyTimeout)
	}
	if cfg.JWKS != "http://jwks.test" {
		t.Errorf("JWKS=%q", cfg.JWKS)
	}
	if cfg.OathkeeperAuthzToken != "callback-token" {
		t.Errorf("Oathkeeper authorization token=%q", cfg.OathkeeperAuthzToken)
	}
	if cfg.OathkeeperURL != "http://oathkeeper-proxy.default:4455" {
		t.Errorf("Oathkeeper URL=%q", cfg.OathkeeperURL)
	}
	if cfg.ServiceAccountTokenPath != "/var/run/secrets/tokens/token" {
		t.Errorf("service account token path=%q", cfg.ServiceAccountTokenPath)
	}
	if cfg.FliptEnabled ||
		cfg.FliptURL != "http://oathkeeper-proxy.default:4455/flipt" ||
		cfg.FliptEnvironment != "local" ||
		cfg.FliptNamespace != "default" ||
		cfg.FliptTestEnvironment != "test" ||
		cfg.FliptUpdateInterval != 30*time.Second ||
		cfg.FliptRequestTimeout != 5*time.Second ||
		cfg.FliptStartupTimeout != 3*time.Second ||
		cfg.FliptManagementURL != "http://oathkeeper-proxy.default:4455/flipt-management" {
		t.Errorf("unexpected Flipt defaults: %+v", cfg)
	}
	t.Setenv("API_FLIPT_ENABLED", "true")
	for variable, test := range map[string]struct {
		validURL string
		field    string
	}{
		"API_OATHKEEPER_URL":       {validURL: "http://oathkeeper.test", field: "OathkeeperURL"},
		"API_FLIPT_URL":            {validURL: "http://flipt.test/evaluation", field: "FliptURL"},
		"API_FLIPT_MANAGEMENT_URL": {validURL: "http://flipt.test/management", field: "FliptManagementURL"},
	} {
		for _, invalidURL := range []string{"not-a-url", "http://:4455", "http://user:secret@gateway.test", "http://gateway.test?route=1", "http://gateway.test#fragment"} {
			t.Setenv(variable, invalidURL)
			if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), test.field) {
				t.Errorf("invalid %s=%q error=%v", variable, invalidURL, err)
			}
		}
		t.Setenv(variable, test.validURL+"/")
	}
	canonicalFlipt, err := loadConfig()
	if err != nil {
		t.Fatalf("canonical Flipt configuration: %v", err)
	}
	if canonicalFlipt.OathkeeperURL != "http://oathkeeper.test" || canonicalFlipt.FliptURL != "http://flipt.test/evaluation" ||
		canonicalFlipt.FliptManagementURL != "http://flipt.test/management" {
		t.Errorf("canonical Flipt URLs: %+v", canonicalFlipt)
	}
	for variable, field := range map[string]string{
		"API_FLIPT_ENVIRONMENT": "FliptEnvironment",
		"API_FLIPT_NAMESPACE":   "FliptNamespace",
	} {
		t.Setenv(variable, "   ")
		if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), field) {
			t.Errorf("blank %s error=%v", variable, err)
		}
		t.Setenv(variable, "default")
	}
	for _, value := range []string{"", "   ", "default"} {
		t.Setenv("API_FLIPT_TEST_ENVIRONMENT", value)
		if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "FliptTestEnvironment") {
			t.Errorf("unsafe Flipt test environment %q error=%v", value, err)
		}
	}
	t.Setenv("API_FLIPT_TEST_ENVIRONMENT", "test")
	t.Setenv("API_FLIPT_ENABLED", "false")
	t.Setenv("API_OATHKEEPER_URL", "")
	t.Setenv("API_FLIPT_URL", "")
	t.Setenv("API_FLIPT_MANAGEMENT_URL", "")
	t.Setenv("API_FLIPT_ENVIRONMENT", "   ")
	t.Setenv("API_FLIPT_NAMESPACE", "   ")
	if _, err := loadConfig(); err != nil {
		t.Errorf("disabled Flipt rejected dormant provider configuration: %v", err)
	}
	for variable, field := range map[string]string{
		"API_OATHKEEPER_URL":       "OathkeeperURL",
		"API_FLIPT_MANAGEMENT_URL": "FliptManagementURL",
	} {
		for _, invalidURL := range []string{
			"http://user:secret@gateway.test",
			"http://gateway.test?route=1",
			"http://gateway.test#fragment",
			"http://:4455",
		} {
			t.Setenv(variable, invalidURL)
			if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), field) {
				t.Errorf("disabled Flipt invalid %s=%q error=%v", variable, invalidURL, err)
			}
		}
		t.Setenv(variable, "")
	}
	t.Setenv("API_OATHKEEPER_URL", "http://oathkeeper.test")
	t.Setenv("API_FLIPT_URL", "http://flipt.test/evaluation")
	t.Setenv("API_FLIPT_MANAGEMENT_URL", "http://flipt.test/management")
	t.Setenv("API_FLIPT_ENVIRONMENT", "local")
	t.Setenv("API_FLIPT_NAMESPACE", "default")
	for _, value := range []string{"true", "false"} {
		t.Setenv("API_SCORING_ENGINE_ENABLED", value)
		loaded, err := loadConfig()
		if err != nil || loaded.ScoringEngineEnabled != (value == "true") {
			t.Fatalf("scoring setting %q: config=%v error=%v", value, loaded.ScoringEngineEnabled, err)
		}
	}
	t.Setenv("API_SCORING_ENGINE_ENABLED", "invalid")
	if _, err := loadConfig(); err == nil {
		t.Fatal("malformed scoring setting accepted")
	}
	if err := os.Unsetenv("API_SCORING_ENGINE_ENABLED"); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(); err == nil {
		t.Fatal("missing scoring setting accepted")
	}
	t.Setenv("API_SCORING_ENGINE_ENABLED", "false")
	t.Setenv("API_JWKS", "")
	if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "JWKS") {
		t.Errorf("missing JWKS configuration error=%v", err)
	}
	for _, rawURL := range []string{
		"not-a-url", "/relative", "ftp://host/jwks.json", "http://:4456",
		"http://user:pass@host/jwks.json", "http://host/jwks.json?x=1", "http://host/jwks.json?",
		"http://host/jwks.json#f", "http://host/jwks.json#",
	} {
		t.Setenv("API_JWKS", rawURL)
		if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "JWKS") {
			t.Errorf("JWKS URL %q error=%v", rawURL, err)
		}
	}
	t.Setenv("API_JWKS", "http://oathkeeper-api.tdk-dev-oathkeeper:4456/.well-known/jwks.json")
	jwksConfig, err := loadConfig()
	if err != nil {
		t.Fatalf("load explicit JWKS configuration: %v", err)
	}
	if jwksConfig.JWKS != "http://oathkeeper-api.tdk-dev-oathkeeper:4456/.well-known/jwks.json" {
		t.Errorf("JWKS=%q", jwksConfig.JWKS)
	}
	t.Setenv("API_JWKS", "http://jwks.test/")
	canonicalJWKS, err := loadConfig()
	if err != nil {
		t.Fatalf("load canonical JWKS configuration: %v", err)
	}
	if canonicalJWKS.JWKS != "http://jwks.test" {
		t.Errorf("JWKS=%q", canonicalJWKS.JWKS)
	}
	t.Setenv("API_JWKS", "http://jwks.test")
	t.Setenv("API_OATHKEEPER_AUTHZ_TOKEN", "")
	if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "OathkeeperAuthzToken") {
		t.Errorf("missing Oathkeeper callback credential error=%v", err)
	}
	t.Setenv("API_OATHKEEPER_AUTHZ_TOKEN", "callback-token")
	t.Setenv("API_JWT_ISSUER", "https://issuer.example.test/")
	issuerConfig, err := loadConfig()
	if err != nil {
		t.Fatalf("load issuer configuration: %v", err)
	}
	if issuerConfig.JWTIssuer != "https://issuer.example.test/" {
		t.Errorf("JWT issuer=%q", issuerConfig.JWTIssuer)
	}
	t.Setenv("API_MAX_TOKEN_AGE", "0s")
	if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "MaxTokenAge") {
		t.Errorf("invalid maximum token age error=%v", err)
	}
	t.Setenv("API_MAX_TOKEN_AGE", "24h")
	t.Setenv("API_VALKEY_TIMEOUT", "0s")
	if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "ValkeyTimeout") {
		t.Errorf("invalid Valkey timeout error=%v", err)
	}
}

func TestLoadConfigRejectsInvalidBranch(t *testing.T) {
	for key, value := range map[string]string{
		"API_SCORING_ENGINE_ENABLED": "false",
		"API_JWKS":                   "http://jwks.test",
		"API_OATHKEEPER_AUTHZ_TOKEN": "callback-token",
		"API_VALKEY_URL":             "redis://valkey:6379",
		"API_POSTGRES_HOST":          "localhost",
		"API_POSTGRES_DATABASE":      "tadoku",
		"API_POSTGRES_USER":          "tadoku",
		"API_POSTGRES_PASSWORD":      "synthetic",
		"API_POSTGRES_SSLMODE":       "disable",
	} {
		t.Setenv(key, value)
	}

	for _, raw := range []string{
		"tadoku/prod",
		"tadoku",
		"bad_key/id",
		"e2e/",
		"/test",
		"TADOKU/prod",
	} {
		t.Run(raw, func(t *testing.T) {
			t.Setenv("API_BRANCH", raw)
			if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "API_BRANCH") {
				t.Errorf("API_BRANCH %q: error=%v", raw, err)
			}
		})
	}
}
