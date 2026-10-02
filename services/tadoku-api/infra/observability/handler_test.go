package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestTenantHandlerPreservesContextAcrossDerivedLoggers(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(NewTenantHandler(slog.NewJSONHandler(&output, nil))).With("component", "worker")
	key, err := tenant.Parse("e2e/logs-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	ctx := tenant.WithKey(t.Context(), key)

	logger.InfoContext(ctx, "scoped event", "job_id", 7)
	logger.WithGroup("detail").InfoContext(ctx, "grouped event", "operation", "invalidate")
	logger.InfoContext(context.Background(), "unscoped event")

	decoder := json.NewDecoder(&output)
	var scoped, grouped, unscoped map[string]any
	for _, event := range []*map[string]any{&scoped, &grouped, &unscoped} {
		if err := decoder.Decode(event); err != nil {
			t.Fatal(err)
		}
	}
	if scoped["tenant"] != key.String() || scoped["component"] != "worker" || scoped["job_id"] != float64(7) {
		t.Errorf("scoped event=%v", scoped)
	}
	detail, ok := grouped["detail"].(map[string]any)
	if !ok ||
		detail["tenant"] != key.String() ||
		detail["operation"] != "invalidate" ||
		grouped["component"] != "worker" {
		t.Errorf("derived grouped event=%v", grouped)
	}
	if _, exists := unscoped["tenant"]; exists {
		t.Errorf("unscoped event has tenant=%v", unscoped["tenant"])
	}
}
