package postgresconfig

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

const defaultPort = 5432

var allowedSSLModes = map[string]bool{
	"disable": true, "allow": true, "prefer": true, "require": true,
	"verify-ca": true, "verify-full": true,
}

// Config is the complete PostgreSQL connection configuration. Password is
// intentionally omitted from String and Redact output.
type Config struct {
	Host, Database, User, Password, SSLMode string
	Port                                    uint16
	legacyURL                               string
}

// Load reads PREFIX_HOST, PORT, DATABASE, USER, PASSWORD, and SSLMODE. During
// the compatibility release, legacyName is accepted only when no individual
// field is present.
func Load(prefix, legacyName string) (Config, error) {
	keys := []string{"HOST", "PORT", "DATABASE", "USER", "PASSWORD", "SSLMODE"}
	values := make(map[string]string, len(keys))
	individualPresent := false
	for _, key := range keys {
		value, present := os.LookupEnv(prefix + "_" + key)
		values[key] = value
		individualPresent = individualPresent || present
	}
	legacy, legacyPresent := os.LookupEnv(legacyName)
	if legacyPresent && individualPresent {
		return Config{}, fmt.Errorf("postgres configuration mixes deprecated %s with individual fields", legacyName)
	}
	if legacyPresent {
		if strings.TrimSpace(legacy) == "" {
			return Config{}, fmt.Errorf("%s must not be empty", legacyName)
		}
		return Config{legacyURL: legacy}, nil
	}

	missing := make([]string, 0)
	for _, key := range []string{"HOST", "DATABASE", "USER", "PASSWORD", "SSLMODE"} {
		if strings.TrimSpace(values[key]) == "" {
			missing = append(missing, prefix+"_"+key)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing postgres configuration: %s", strings.Join(missing, ", "))
	}
	port := defaultPort
	if values["PORT"] != "" {
		parsed, err := strconv.Atoi(values["PORT"])
		if err != nil || parsed < 1 || parsed > 65535 {
			return Config{}, fmt.Errorf("%s_PORT must be an integer between 1 and 65535", prefix)
		}
		port = parsed
	}
	if !allowedSSLModes[values["SSLMODE"]] {
		return Config{}, fmt.Errorf("%s_SSLMODE is invalid", prefix)
	}
	return Config{Host: values["HOST"], Port: uint16(port), Database: values["DATABASE"], User: values["USER"], Password: values["PASSWORD"], SSLMode: values["SSLMODE"]}, nil
}

func (c Config) URL() string {
	if c.legacyURL != "" {
		return c.legacyURL
	}
	u := &url.URL{Scheme: "postgres", User: url.UserPassword(c.User, c.Password), Host: net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port))), Path: "/" + c.Database}
	query := u.Query()
	query.Set("sslmode", c.SSLMode)
	u.RawQuery = query.Encode()
	return u.String()
}

func (c Config) ConnConfig() (*pgx.ConnConfig, error) {
	config, err := pgx.ParseConfig(c.URL())
	if err != nil {
		return nil, fmt.Errorf("parse postgres configuration: %s", c.Redact(err))
	}
	return config, nil
}

func (c Config) String() string { return "postgres configuration (credentials redacted)" }

func (c Config) Redact(value any) string {
	result := fmt.Sprint(value)
	for _, secret := range []string{c.Password, c.legacyURL} {
		if secret != "" {
			result = strings.ReplaceAll(result, secret, "[REDACTED]")
		}
	}
	if parsed, err := url.Parse(c.legacyURL); err == nil && parsed.User != nil {
		if password, ok := parsed.User.Password(); ok && password != "" {
			result = strings.ReplaceAll(result, password, "[REDACTED]")
		}
	}
	return result
}
