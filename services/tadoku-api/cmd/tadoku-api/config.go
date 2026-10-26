package main

import (
	"fmt"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	fliptclient "github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type config struct {
	Production              bool   `envconfig:"production" default:"true"`
	Branch                  string `envconfig:"branch"`
	deployment              tenant.Deployment
	ScoringEngineEnabled    bool   `envconfig:"scoring_engine_enabled" required:"true"`
	Port                    int    `validate:"gt=0,lte=65535" default:"8000"`
	MetricsPort             int    `validate:"gt=0,lte=65535" envconfig:"metrics_port" default:"9090"`
	ServiceName             string `validate:"required" envconfig:"service_name" default:"tadoku-api"`
	JWKS                    string `validate:"required"`
	JWTIssuer               string `envconfig:"jwt_issuer"`
	OathkeeperAuthzToken    string `validate:"required" envconfig:"oathkeeper_authz_token"`
	OathkeeperURL           string `envconfig:"oathkeeper_url" default:"http://oathkeeper-proxy.default:4455"`
	ServiceAccountTokenPath string `validate:"required" envconfig:"service_account_token_path" default:"/var/run/secrets/tokens/token"`

	FliptEnabled         bool          `envconfig:"flipt_enabled" default:"false"`
	FliptURL             string        `envconfig:"flipt_url" default:"http://oathkeeper-proxy.default:4455/flipt"`
	FliptEnvironment     string        `envconfig:"flipt_environment" default:"local"`
	FliptTestEnvironment string        `envconfig:"flipt_test_environment" default:"test"`
	FliptNamespace       string        `envconfig:"flipt_namespace" default:"default"`
	FliptUpdateInterval  time.Duration `validate:"gte=1s" envconfig:"flipt_update_interval" default:"30s"`
	FliptRequestTimeout  time.Duration `validate:"gte=1s" envconfig:"flipt_request_timeout" default:"5s"`
	FliptStartupTimeout  time.Duration `validate:"gt=0" envconfig:"flipt_startup_timeout" default:"3s"`

	IdentityCookieDomain      string   `envconfig:"identity_cookie_domain" default:"tadoku.app"`
	IdentityTrustedProxyCIDRs []string `envconfig:"identity_trusted_proxy_cidrs"`
	identityTrustedProxies    []netip.Prefix

	PostgresMaxConnections int32                 `validate:"gt=0,lte=32" envconfig:"postgres_max_connections" default:"4"`
	Postgres               postgresconfig.Config `ignored:"true"`
	ValkeyURL              string                `validate:"required" envconfig:"valkey_url"`
	ValkeyTimeout          time.Duration         `validate:"gt=0" envconfig:"valkey_timeout" default:"1s"`

	DialTimeout           time.Duration `validate:"gt=0" envconfig:"dial_timeout" default:"3s"`
	MaxTokenAge           time.Duration `validate:"gt=0" envconfig:"max_token_age" default:"24h"`
	ResponseHeaderTimeout time.Duration `validate:"gt=0" envconfig:"response_header_timeout" default:"10s"`
	RequestTimeout        time.Duration `validate:"gt=0" envconfig:"request_timeout" default:"30s"`
	IdleTimeout           time.Duration `validate:"gt=0" envconfig:"idle_timeout" default:"30s"`
	ShutdownTimeout       time.Duration `validate:"gt=0" envconfig:"shutdown_timeout" default:"10s"`
}

func loadConfig() (config, error) {
	cfg := config{}
	if err := envconfig.Process("API", &cfg); err != nil {
		return config{}, fmt.Errorf("load config: %w", err)
	}

	deployment, err := tenant.ParseDeployment(cfg.Branch)
	if err != nil {
		return config{}, fmt.Errorf("validate config: API_BRANCH: %w", err)
	}
	cfg.deployment = deployment

	if err := validator.New().Struct(cfg); err != nil {
		return config{}, fmt.Errorf("validate config: %w", err)
	}
	if cfg.FliptEnabled {
		if strings.TrimSpace(cfg.FliptEnvironment) == "" {
			return config{}, fmt.Errorf("validate config: FliptEnvironment is required")
		}
		if strings.TrimSpace(cfg.FliptNamespace) == "" {
			return config{}, fmt.Errorf("validate config: FliptNamespace is required")
		}

		_, targetsErr := fliptclient.NewTargets(
			cfg.FliptEnvironment,
			cfg.FliptNamespace,
			cfg.FliptTestEnvironment,
		)
		if targetsErr != nil {
			return config{}, fmt.Errorf("validate config: FliptTestEnvironment: %w", targetsErr)
		}

		cfg.FliptURL, err = providerBaseURL("FliptURL", cfg.FliptURL)
		if err != nil {
			return config{}, err
		}
	}

	if cfg.OathkeeperURL != "" {
		cfg.OathkeeperURL, err = providerBaseURL("OathkeeperURL", cfg.OathkeeperURL)
		if err != nil {
			return config{}, err
		}
	}

	cfg.JWKS, err = providerBaseURL("JWKS", cfg.JWKS)
	if err != nil {
		return config{}, err
	}

	if cfg.IdentityCookieDomain != "tadoku.app" && cfg.IdentityCookieDomain != "tadoku.dev.lab" {
		return config{}, fmt.Errorf("validate config: IdentityCookieDomain must be tadoku.app or tadoku.dev.lab")
	}
	for _, raw := range cfg.IdentityTrustedProxyCIDRs {
		prefix, parseErr := netip.ParsePrefix(raw)
		if parseErr != nil {
			return config{}, fmt.Errorf("validate config: IdentityTrustedProxyCIDRs must contain IP prefixes")
		}
		cfg.identityTrustedProxies = append(cfg.identityTrustedProxies, prefix)
	}

	cfg.Postgres, err = postgresconfig.Load("API_POSTGRES", "API_POSTGRES_URL")
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}

func providerBaseURL(name, rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Hostname() == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") ||
		parsedURL.User != nil || parsedURL.RawQuery != "" || parsedURL.ForceQuery || strings.Contains(rawURL, "#") {
		return "", fmt.Errorf("validate config: %s must be an HTTP(S) URL without credentials, query or fragment", name)
	}
	return strings.TrimRight(rawURL, "/"), nil
}
