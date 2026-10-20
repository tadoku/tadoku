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
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/google/uuid"
	kratosapi "github.com/ory/kratos-client-go"
	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"go.yaml.in/yaml/v3"
)

func TestEmailGatewayCookieAndBearer(t *testing.T) {
	api.reset(t, "")
	resetKratosAfter(t, api)

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

	cookie := gatewayAdminCookie(t, key, true)
	readerCookie := gatewayAdminCookie(t, key, false)
	backend := httptest.NewUnstartedServer(nil)
	defer backend.Close()

	proxyURL, jwksURL := startEmailGateway(t, "http://"+backend.Listener.Addr().String(), cookie)
	deployment, _ := tenant.ParseDeployment(key.String())

	handler, _, _, err := newTestRouterWithLeaderboardService(t.Context(), api.db.AppPool, api.db.AppPool,
		api.keto, api.kratos, slog.New(slog.NewTextHandler(io.Discard, nil)), false, nil, deployment,
		func(e *routerEndpoints) { e.jwks = jwksURL })
	if err != nil {
		t.Fatal(err)
	}

	var active atomic.Value
	active.Store(http.HandlerFunc(handler.ServeHTTP))
	backend.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		active.Load().(http.HandlerFunc).ServeHTTP(w, request)
	})
	backend.Start()

	hosts := []string{"tadoku.dev.lab"}
	forwardedProto := ""
	if os.Getenv("TADOKU_TEST_EMAIL_GATEWAY_RENDER") != "" {
		hosts = []string{"tadoku.app", "r2-0123abcd.preview.tadoku.app"}
		forwardedProto = "https"
	}
	host := hosts[0]
	branch := "gateway-0123abcd"
	client := &http.Client{Timeout: 3 * time.Second}

	check := func(name, path, authorization, cookie string, want int) {
		t.Helper()
		req, err := http.NewRequestWithContext(t.Context(), "GET", proxyURL+"/api/internal"+path, nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Host = host
		req.Header.Set("X-Forwarded-Proto", forwardedProto)
		req.Header.Set("X-Dev-Branch", branch)
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
		if path == "/authz/current-user/role" && want == 200 && string(body) != "{\"role\":\"guest\"}\n" {
			t.Errorf("token on guest-readable route did not stay guest: %s", body)
		}
		if want == 200 && path == "/identity/session" && !bytes.Contains(body, []byte("account_app")) {
			t.Errorf("%s did not return the actual identity session", name)
		}
		if want == 200 && strings.HasPrefix(path, "/emails") && !bytes.Contains(body, []byte(id.String())) {
			t.Errorf("%s did not return owned email", name)
		}
	}

	for _, host = range hosts {
		check("public identity guest", "/identity/session", "", "", 200)
		check("public identity unrelated", "/identity/session", "", "analytics=unrelated", 200)
		check("public identity valid cookie", "/identity/session", "", cookie, 200)
		check("bearer list", "/emails", "Bearer "+raw, "", 200)
		check("bearer unrelated cookie", "/emails", "Bearer "+raw, "analytics=unrelated", 200)
		check("bearer expired Kratos", "/emails", "Bearer "+raw, "ory_kratos_session=expired", 200)
		check("bearer real Kratos cookie", "/emails", "Bearer "+raw, cookie, 200)
		check("bearer mixed cookies", "/emails", "Bearer "+raw, "analytics=unrelated; tadoku_session=invalid; "+
			"tadoku_session_tadoku_gateway-0123abcd=invalid; "+cookie, 200)
		check("invalid token admin cookie", "/emails", "Bearer tdk_at_invalid", cookie, 401)
		check("unrelated guest", "/emails", "", "analytics=unrelated", 401)
		check("expired Kratos guest", "/emails", "", "ory_kratos_session=expired", 401)
		check("bearer outside email stays guest", "/authz/current-user/role", "Bearer "+raw, "analytics=unrelated", 200)
		check("bearer protected route with cookie denied", "/profile/users", "Bearer "+raw, "analytics=unrelated", 401)
		check("bearer detail", "/emails/"+id.String(), "Bearer "+raw, "", 200)
		check("bearer reader cookie precedence", "/emails", "Bearer "+raw, readerCookie, 200)
		check("reader cookie denied", "/emails", "", readerCookie, 403)
		check("forged bearer admin cookie replaced", "/emails", "Bearer invalid", cookie, 200)
		check("cookie admin", "/emails", "", cookie, 200)
		check("anonymous", "/emails", "", "", 401)
		check("forged jwt", "/emails", "Bearer invalid", "", 401)
		check("token outside email", "/authz/current-user/role", "Bearer "+raw, "", 401)
		check("cookie detail", "/emails/"+id.String(), "", cookie, 200)
	}
	providerClient := api.kratos.PublicClient()
	request, _ := http.NewRequestWithContext(t.Context(), "GET", "http://kratos-public.test/self-service/logout/browser", nil)
	request.Header.Set("Cookie", cookie)
	logout, err := providerClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var logoutView struct {
		LogoutURL string `json:"logout_url"`
	}
	err = json.NewDecoder(logout.Body).Decode(&logoutView)
	logout.Body.Close()
	if err != nil || logoutView.LogoutURL == "" {
		t.Fatalf("real public logout init: %v", err)
	}
	request, _ = http.NewRequestWithContext(t.Context(), "GET", logoutView.LogoutURL, nil)
	request.Header.Set("Cookie", cookie)
	logout, err = providerClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	logout.Body.Close()
	request, _ = http.NewRequestWithContext(t.Context(), "GET", "http://kratos-public.test/sessions/whoami", nil)
	request.Header.Set("Cookie", cookie)
	who, err := providerClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	who.Body.Close()
	if who.StatusCode != 401 {
		t.Fatalf("disposable old session status=%d", who.StatusCode)
	}
	check("real revoked Kratos cookie guest", "/emails", "", cookie, 401)
	check("bearer real revoked Kratos cookie", "/emails", "Bearer "+raw, cookie, 200)
	if err := api.kratos.ExpireOwnedSessions(t.Context()); err != nil {
		t.Fatal(err)
	}
	request, _ = http.NewRequestWithContext(t.Context(), "GET", "http://kratos-public.test/sessions/whoami", nil)
	request.Header.Set("Cookie", readerCookie)
	who, err = providerClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	who.Body.Close()
	if who.StatusCode != 401 {
		t.Fatalf("disposable expired session status=%d", who.StatusCode)
	}
	check("real expired Kratos cookie guest", "/emails", "", readerCookie, 401)
	check("bearer real expired Kratos cookie", "/emails", "Bearer "+raw, readerCookie, 200)
	other, _ := tenant.ParseDeployment("tadoku/other-0123abcd")
	otherHandler, _, _, err := newTestRouterWithLeaderboardService(t.Context(), api.db.AppPool, api.db.AppPool,
		api.keto, api.kratos, slog.New(slog.NewTextHandler(io.Discard, nil)), false, nil, other,
		func(e *routerEndpoints) { e.jwks = jwksURL })
	if err != nil {
		t.Fatal(err)
	}
	active.Store(http.HandlerFunc(otherHandler.ServeHTTP))
	branch = "other-0123abcd"
	check("token wrong deployment", "/emails", "Bearer "+raw, "analytics=unrelated", 421)
	active.Store(http.HandlerFunc(handler.ServeHTTP))
	branch = "gateway-0123abcd"
	_, err = api.db.Pool.Exec(t.Context(),
		"update access_tokens set revoked_at=now() where token_hash=$1", hex.EncodeToString(hash[:]),
	)
	if err != nil {
		t.Fatal(err)
	}
	check("revoked token", "/emails", "Bearer "+raw, "", 401)
	if os.Getenv("TADOKU_TEST_EMAIL_GATEWAY_RENDER") != "" {
		host = hosts[0]
		forwardedProto = ""
		check("plain HTTP refused", "/emails", "Bearer "+raw, "", 404)
	}
}

func gatewayAdminCookie(t *testing.T, key tenant.Key, admin bool) string {
	t.Helper()

	password := rand.Text() + "!a1"
	email := uuid.NewString() + "@example.test"
	credentials := &kratosapi.IdentityWithCredentials{}
	rawCredentials, err := json.Marshal(map[string]any{
		"password": map[string]any{"config": map[string]string{"password": password}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(rawCredentials, credentials); err != nil {
		t.Fatal(err)
	}

	identityAPI := api.kratos.Client().IdentityApi
	actor, _, err := identityAPI.CreateIdentity(t.Context()).CreateIdentityBody(kratosapi.CreateIdentityBody{
		SchemaId:    "user",
		Traits:      map[string]interface{}{"email": email, "display_name": "Gateway Account"},
		Credentials: credentials,
	}).Execute()
	if err != nil {
		t.Fatal(err)
	}

	roles := ketoclient.NewClient(api.keto.ReadURL(), api.keto.WriteURL())
	if admin {
		err = roles.AddRelation(t.Context(), "app", key.String(), "admins", ketoclient.Subject{ID: actor.Id})
		if err != nil {
			t.Fatal(err)
		}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	client := api.kratos.PublicClient()
	client.Jar = jar
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	req, err := http.NewRequestWithContext(
		t.Context(), "GET", "http://kratos-public.test/self-service/login/browser", nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	var flow struct {
		ID string `json:"id"`
		UI struct {
			Nodes []struct {
				Attributes struct {
					Name  string
					Value any
				}
			}
		}
	}

	err = json.NewDecoder(response.Body).Decode(&flow)
	response.Body.Close()
	if err != nil || response.StatusCode != 200 || flow.ID == "" {
		t.Fatalf("initialize real Kratos browser login: status=%d error=%v", response.StatusCode, err)
	}

	var csrf any
	for _, node := range flow.UI.Nodes {
		if node.Attributes.Name == "csrf_token" {
			csrf = node.Attributes.Value
		}
	}

	body, err := json.Marshal(map[string]any{
		"method":     "password",
		"identifier": email,
		"password":   password,
		"csrf_token": csrf,
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequestWithContext(
		t.Context(), "POST", "http://kratos-public.test/self-service/login?flow="+flow.ID, bytes.NewReader(body),
	)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	response, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatalf("real Kratos browser login status=%d", response.StatusCode)
	}

	u, _ := url.Parse("http://kratos-public.test")
	for _, cookie := range jar.Cookies(u) {
		if cookie.Name == "ory_kratos_session" {
			return cookie.String()
		}
	}

	t.Fatal("real Kratos browser login returned no session cookie")
	return ""
}

func startEmailGateway(t *testing.T, backend, cookie string) (string, string) {
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
	serve["prometheus"] = map[string]any{"host": "127.0.0.1", "port": port()}

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

	command.Env = append(os.Environ(),
		`AUTHENTICATORS_COOKIE_SESSION_CONFIG_ADDITIONAL_HEADERS={"Authorization":"Bearer `+callbackToken+`"}`,
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

		if strings.Contains(logs.String(), "tdk_at_") || strings.Contains(logs.String(), cookie) ||
			strings.Contains(logs.String(), callbackToken) {
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
