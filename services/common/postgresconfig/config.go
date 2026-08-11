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
}

// Load reads PREFIX_HOST, PORT, DATABASE, USER, PASSWORD, and SSLMODE. The
// removed legacy URL variable is rejected explicitly so regressions fail closed.
func Load(prefix, legacyName string) (Config, error) {
	keys := []string{"HOST", "PORT", "DATABASE", "USER", "PASSWORD", "SSLMODE"}
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		values[key] = os.Getenv(prefix + "_" + key)
	}
	_, legacyPresent := os.LookupEnv(legacyName)
	if legacyPresent {
		return Config{}, fmt.Errorf("%s is no longer supported; use individual postgres fields", legacyName)
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
	for _, secret := range []string{c.Password} {
		if secret != "" {
			result = strings.ReplaceAll(result, secret, "[REDACTED]")
		}
	}
	return result
}
