package e2e_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/google/uuid"
	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"go.yaml.in/yaml/v3"
	"golang.org/x/crypto/bcrypt"
)

func TestEmailGatewayCookieAndBearer(t *testing.T) {
	api.reset(t, "")

	key, _ := tenant.Parse("tadoku/gateway-0123abcd")
	_, err := api.db.Pool.Exec(t.Context(), "insert into tenants(key,kind) values($1,'test')", key.String())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := api.db.Pool.Exec(ctx, "delete from tenants where key=$1", key.String()); err != nil {
			t.Error(err)
		}
	})

	id := writeEmail(t, tenant.WithKey(t.Context(), key), emailService(), recoveryMail())

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}

	raw := "tdk_at_tadoku_gateway-0123abcd_" + base64.RawURLEncoding.EncodeToString(secret)
	hash := sha256.Sum256([]byte(raw))
	if _, err := api.db.Pool.Exec(t.Context(), `insert into access_tokens(tenant,name,token_hash,scopes,expires_at)
  values($1,'proxy', $2,array['email:read'],now()+interval '1 hour')`,
		key.String(), hex.EncodeToString(hash[:])); err != nil {
		t.Fatal(err)
	}

	if err := flipt.SetIdentityV2(key, true); err != nil {
		t.Fatal(err)
	}
	var cookie string
	backend := httptest.NewUnstartedServer(nil)
	defer backend.Close()

	proxyURL, jwksURL := startEmailGateway(t, "http://"+backend.Listener.Addr().String(), &cookie)
	deployment, _ := tenant.ParseDeployment(key.String())

	handler, _, _, err := newTestRouterWithLeaderboardService(t.Context(), api.db.AppPool, api.db.AppPool,
		api.keto, api.kratos, slog.New(slog.NewTextHandler(io.Discard, nil)), false, nil, deployment,
		func(e *routerEndpoints) { e.jwks = jwksURL })
	if err != nil {
		t.Fatal(err)
	}

	backend.Config.Handler = handler
	backend.Start()

	hosts := []string{"tadoku.dev.lab"}
	forwardedProto := ""
	if os.Getenv("TADOKU_TEST_EMAIL_GATEWAY_RENDER") != "" {
		hosts = []string{"tadoku.app", "r2-0123abcd.preview.tadoku.app"}
		forwardedProto = "https"
	}
	host := hosts[0]
	client := &http.Client{Timeout: 3 * time.Second}
	cookie = gatewayAdminCookie(t, key, proxyURL, host, forwardedProto)

	check := func(name, path, authorization, cookie string, want int) {
		t.Helper()
		req, err := http.NewRequestWithContext(t.Context(), "GET", proxyURL+"/api/internal"+path, nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Host = host
		req.Header.Set("X-Forwarded-Proto", forwardedProto)
		req.Header.Set("X-Dev-Branch", "gateway-0123abcd")
		req.Header.Set("Authorization", authorization)
		req.Header.Set("Cookie", cookie)

		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("host=%s proto=%s case=%s status=%d want=%d", host, forwardedProto, name, response.StatusCode, want)
		if response.StatusCode != want {
			t.Errorf("%s %s: status=%d want=%d body=%s", host, name, response.StatusCode, want, body)
		}
		if want == 200 && !bytes.Contains(body, []byte(id.String())) {
			t.Errorf("%s did not return owned email", name)
		}
	}

	for _, host = range hosts {
		check("bearer list", "/emails", "Bearer "+raw, "", 200)
		check("bearer detail", "/emails/"+id.String(), "Bearer "+raw, "", 200)
		check("cookie admin", "/emails", "", cookie, 200)
		check("anonymous", "/emails", "", "", 401)
		check("unrelated cookie", "/emails", "", "unrelated=present", 401)
		check("forged session", "/emails", "", "tadoku_session_tadoku_gateway-0123abcd=invalid", 401)
		check("forged jwt", "/emails", "Bearer invalid", "", 401)
		check("token outside email", "/authz/current-user/role", "Bearer "+raw, "", 401)
		check("cookie detail", "/emails/"+id.String(), "", cookie, 200)
	}

	_, err = api.db.Pool.Exec(t.Context(), `update sessions
		set expires_at = now() - interval '1 hour'
		where tenant = $1`, key.String())
	if err != nil {
		t.Fatal(err)
	}
	check("expired session", "/emails", "", cookie, 401)
	_, err = api.db.Pool.Exec(t.Context(), `update sessions
		set expires_at = now() + interval '1 hour'
		where tenant = $1`, key.String())
	if err != nil {
		t.Fatal(err)
	}

	logout, err := http.NewRequestWithContext(t.Context(), "POST", proxyURL+"/api/internal/identity/logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	logout.Host = host
	logout.Header.Set("X-Forwarded-Proto", forwardedProto)
	logout.Header.Set("X-Dev-Branch", "gateway-0123abcd")
	logout.Header.Set("Origin", "https://tadoku.dev.lab")
	logout.Header.Set("Content-Type", "application/json")
	logout.Header.Set("Cookie", cookie)
	loggedOut, err := client.Do(logout)
	if err != nil {
		t.Fatal(err)
	}
	loggedOut.Body.Close()
	if loggedOut.StatusCode != http.StatusNoContent {
		t.Fatalf("logout status=%d", loggedOut.StatusCode)
	}

	check("revoked session", "/emails", "", cookie, 401)
	if os.Getenv("TADOKU_TEST_EMAIL_GATEWAY_RENDER") != "" {
		host = hosts[0]
		forwardedProto = ""
		check("plain HTTP refused", "/emails", "Bearer "+raw, "", 404)
	}
}

func gatewayAdminCookie(t *testing.T, key tenant.Key, proxyURL, host, forwardedProto string) string {
	t.Helper()
	userID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte(identityFixturePassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	_, err = api.db.Pool.Exec(t.Context(), `with fixture_user as (
		insert into users (tenant, id, display_name)
		values ($1, $2, 'Gateway Admin')
		returning tenant, id
	)
	insert into accounts (tenant, user_id, email, email_normalized, password_hash)
	select tenant, id, 'gateway@example.test', 'gateway@example.test', $3
	from fixture_user`, key.String(), userID, string(hash))
	if err != nil {
		t.Fatal(err)
	}

	body := `{"email":"gateway@example.test","password":"Fixture-login-482!"}`
	request, err := http.NewRequestWithContext(
		t.Context(), "POST", proxyURL+"/api/internal/identity/login", strings.NewReader(body),
	)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = host
	request.Header.Set("X-Forwarded-Proto", forwardedProto)
	request.Header.Set("X-Dev-Branch", strings.TrimPrefix(key.String(), "tadoku/"))
	request.Header.Set("Origin", "https://tadoku.dev.lab")
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 3 * time.Second}
	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	t.Logf("gateway login completed in %s (client limit %s)", time.Since(started), client.Timeout)

	var view identitySessionView
	if err := json.NewDecoder(response.Body).Decode(&view); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || view.AccountApp != "v2" || view.Account == nil ||
		view.Account.ID != userID.String() || view.Account.DisplayName != "Gateway Admin" ||
		view.Account.Email != "gateway@example.test" {
		t.Fatalf(
			"gateway login: status=%d app=%s account=%+v",
			response.StatusCode, view.AccountApp, view.Account,
		)
	}

	roles := ketoclient.NewClient(api.keto.ReadURL(), api.keto.WriteURL())
	err = roles.AddRelation(t.Context(), "app", key.String(), "admins", ketoclient.Subject{ID: view.Account.ID})
	if err != nil {
		t.Fatal(err)
	}

	cookies := response.Cookies()
	if len(cookies) != 1 || cookies[0].Name != "tadoku_session_tadoku_gateway-0123abcd" {
		t.Fatal("gateway login did not issue the tenant's v2 cookie")
	}
	return cookies[0].Name + "=" + cookies[0].Value
}

func startEmailGateway(t *testing.T, backend string, cookie *string) (string, string) {
	t.Helper()

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	if err := os.Mkdir(configDir, 0700); err != nil {
		t.Fatal(err)
	}

	resolve := func(path string) string {
		t.Helper()
		result, err := runfiles.Rlocation(path)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}

	load := func(path string) []byte {
		t.Helper()
		raw, err := os.ReadFile(resolve(path))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}

	configRaw := load("tadoku/k8s/dev/base/oathkeeper/oathkeeper.yaml")
	rulesRaw := load("tadoku/k8s/dev/base/oathkeeper/rules.yaml")
	binary := "oathkeeper_v26_2_0/oathkeeper"
	if rendered := os.Getenv("TADOKU_TEST_EMAIL_GATEWAY_RENDER"); rendered != "" {
		configRaw, rulesRaw = renderedEmailGateway(t, rendered)
		binary = "oathkeeper_v25_4_0/oathkeeper"
	}

	var config map[string]any
	if err := yaml.Unmarshal(configRaw, &config); err != nil {
		t.Fatal(err)
	}

	port := func() int {
		t.Helper()
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		p := l.Addr().(*net.TCPAddr).Port
		l.Close()
		return p
	}

	proxyPort, apiPort := port(), port()
	serve := config["serve"].(map[string]any)
	serve["proxy"].(map[string]any)["port"] = proxyPort
	serve["proxy"].(map[string]any)["host"] = "127.0.0.1"
	serve["api"] = map[string]any{"host": "127.0.0.1", "port": apiPort}

	authenticators := config["authenticators"].(map[string]any)
	cookieSession := authenticators["cookie_session"].(map[string]any)["config"].(map[string]any)
	cookieSession["check_session_url"] = backend + "/identity/internal/v1/session"
	cookieSession["additional_headers"] = map[string]string{"Authorization": "Bearer " + callbackToken}
	cookieSession["forward_http_headers"] = []string{"Cookie", "X-Dev-Branch"}
	delete(cookieSession, "only")

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	key.Precompute()
	encode := func(n *big.Int) string { return base64.RawURLEncoding.EncodeToString(n.Bytes()) }

	jwk := map[string]string{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": uuid.NewString(),
		"n":   encode(key.N),
		"e":   encode(big.NewInt(int64(key.E))),
		"d":   encode(key.D),
		"p":   encode(key.Primes[0]),
		"q":   encode(key.Primes[1]),
		"dp":  encode(key.Precomputed.Dp),
		"dq":  encode(key.Precomputed.Dq),
		"qi":  encode(key.Precomputed.Qinv),
	}

	write := func(path string, value any) {
		t.Helper()
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}

	jwksPath := filepath.Join(configDir, "jwks.json")
	write(jwksPath, map[string]any{"keys": []any{jwk}})

	mutators := config["mutators"].(map[string]any)
	mutators["id_token"].(map[string]any)["config"].(map[string]any)["jwks_url"] = "file://" + jwksPath

	rulesPath := filepath.Join(configDir, "rules.yaml")
	rules := strings.ReplaceAll(
		string(rulesRaw),
		"http://tadoku-cli-api.tdk-dev-oathkeeper.svc.cluster.local",
		backend,
	)
	rules = strings.ReplaceAll(rules, "http://tadoku-api-internal.tdk-prod-oathkeeper.svc.cluster.local", backend)
	if err := os.WriteFile(rulesPath, []byte(rules), 0600); err != nil {
		t.Fatal(err)
	}

	config["access_rules"].(map[string]any)["repositories"] = []string{"file://" + rulesPath}
	configPath := filepath.Join(configDir, "oathkeeper.json")
	write(configPath, config)

	command := exec.CommandContext(
		t.Context(), resolve(binary), "serve", "--sqa-opt-out", "--config", configPath,
	)

	logs := new(bytes.Buffer)
	command.Stdout = logs
	command.Stderr = logs
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() { _ = command.Wait(); close(done) }()

	t.Cleanup(func() {
		command.Process.Signal(os.Interrupt)

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			command.Process.Kill()
			<-done
		}

		if strings.Contains(logs.String(), "tdk_at_") || (*cookie != "" && strings.Contains(logs.String(), *cookie)) {
			t.Error("Oathkeeper logs exposed token or session cookie material")
		}
	})

	proxyURL := "http://" + net.JoinHostPort("127.0.0.1", fmt.Sprint(proxyPort))
	apiURL := "http://" + net.JoinHostPort("127.0.0.1", fmt.Sprint(apiPort))

	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(apiURL + "/health/ready")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == 200 {
				return proxyURL, apiURL + "/.well-known/jwks.json"
			}
		}

		select {
		case <-done:
			t.Fatalf("Oathkeeper exited before ready: %s", logs.String())
		default:
		}

		time.Sleep(25 * time.Millisecond)
	}

	t.Fatal("Oathkeeper readiness timed out")
	return "", ""
}

func renderedEmailGateway(t *testing.T, path string) ([]byte, []byte) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var config, rules []byte
	for {
		var document struct {
			Kind string
			Data map[string]string
		}
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if document.Kind != "ConfigMap" {
			continue
		}

		if value := document.Data["config.yaml"]; value != "" {
			config = []byte(value)
		}
		if value := document.Data["access-rules.json"]; value != "" {
			rules = []byte(value)
		}
	}
	if len(config) == 0 || len(rules) == 0 {
		t.Fatal("Helm render must contain Oathkeeper config and access rules")
	}
	return config, rules
}
