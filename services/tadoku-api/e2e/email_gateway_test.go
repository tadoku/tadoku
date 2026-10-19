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
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	if _, err := api.db.Pool.Exec(t.Context(), "insert into tenants(key,kind) values($1,'test')", key.String()); err != nil {
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

	cookie := gatewayAdminCookie(t, key)
	backend := httptest.NewUnstartedServer(nil)
	defer backend.Close()
	kratosURL, _ := url.Parse("http://kratos-public.test")
	proxy := httputil.NewSingleHostReverseProxy(kratosURL)
	proxy.Transport = api.kratos.PublicClient().Transport
	provider := httptest.NewServer(proxy)
	defer provider.Close()
	proxyURL, jwksURL := startEmailGateway(t, "http://"+backend.Listener.Addr().String(), provider.URL)
	deployment, _ := tenant.ParseDeployment(key.String())
	handler, _, _, err := newTestRouterWithLeaderboardService(t.Context(), api.db.AppPool, api.db.AppPool,
		api.keto, api.kratos, slog.New(slog.NewTextHandler(io.Discard, nil)), false, nil, deployment,
		func(e *routerEndpoints) { e.jwks = jwksURL })
	if err != nil {
		t.Fatal(err)
	}
	backend.Config.Handler = handler
	backend.Start()

	client := &http.Client{Timeout: 3 * time.Second}
	check := func(name, path, authorization, cookie string, want int) {
		t.Helper()
		req, err := http.NewRequestWithContext(t.Context(), "GET", proxyURL+"/api/internal"+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = "tadoku.dev.lab"
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
		if response.StatusCode != want {
			t.Errorf("%s: status=%d want=%d body=%s", name, response.StatusCode, want, body)
		}
		if want == 200 && !bytes.Contains(body, []byte(id.String())) {
			t.Errorf("%s did not return owned email", name)
		}
	}
	check("bearer list", "/emails", "Bearer "+raw, "", 200)
	check("bearer detail", "/emails/"+id.String(), "Bearer "+raw, "", 200)
	check("cookie admin", "/emails", "", cookie, 200)
	check("anonymous", "/emails", "", "", 401)
	check("forged jwt", "/emails", "Bearer invalid", "", 401)
	check("token outside email", "/authz/current-user/role", "Bearer "+raw, "", 401)
	check("cookie detail", "/emails/"+id.String(), "", cookie, 200)
}

func gatewayAdminCookie(t *testing.T, key tenant.Key) string {
	t.Helper()
	password := rand.Text() + "!a1"
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
		Traits:      map[string]interface{}{"email": "gateway@example.test", "display_name": "Gateway Admin"},
		Credentials: credentials,
	}).Execute()
	if err != nil {
		t.Fatal(err)
	}
	roles := ketoclient.NewClient(api.keto.ReadURL(), api.keto.WriteURL())
	err = roles.AddRelation(t.Context(), "app", key.String(), "admins", ketoclient.Subject{ID: actor.Id})
	if err != nil {
		t.Fatal(err)
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := api.kratos.PublicClient()
	client.Jar = jar
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(t.Context(), "GET", "http://kratos-public.test/self-service/login/browser", nil)
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
		"identifier": "gateway@example.test",
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

func startEmailGateway(t *testing.T, backend, provider string) (string, string) {
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
	var config map[string]any
	if err := yaml.Unmarshal(load("tadoku/k8s/dev/base/oathkeeper/oathkeeper.yaml"), &config); err != nil {
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
	cookieSession["check_session_url"] = provider + "/sessions/whoami"
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
		string(load("tadoku/k8s/dev/base/oathkeeper/rules.yaml")),
		"http://tadoku-cli-api.tdk-dev-oathkeeper.svc.cluster.local",
		backend,
	)
	if err := os.WriteFile(rulesPath, []byte(rules), 0600); err != nil {
		t.Fatal(err)
	}
	config["access_rules"].(map[string]any)["repositories"] = []string{"file://" + rulesPath}
	configPath := filepath.Join(configDir, "oathkeeper.json")
	write(configPath, config)
	command := exec.CommandContext(
		t.Context(), resolve("oathkeeper_v26_2_0/oathkeeper"), "serve", "--sqa-opt-out", "--config", configPath,
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
		if strings.Contains(logs.String(), "tdk_at_") {
			t.Error("Oathkeeper logs exposed access token material")
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
