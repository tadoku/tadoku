package http

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

func TestEmailReadHandlersLogOperationErrors(t *testing.T) {
	t.Parallel()

	var logs strings.Builder
	server := server{
		application: app.New(app.Dependencies{Permissions: permissions.NewKetoChecker(nil)}),
		logger:      slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}
	ctx := context.Background()

	if _, err := server.EmailList(ctx, openapi.EmailListRequestObject{}); err == nil {
		t.Fatal("EmailList error=nil, want application failure")
	}
	if _, err := server.EmailGet(ctx, openapi.EmailGetRequestObject{Id: uuid.Nil}); err == nil {
		t.Fatal("EmailGet error=nil, want application failure")
	}

	for _, want := range []string{
		`msg="list emails rejected"`,
		`msg="find email rejected"`,
	} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("log %q does not contain %q", logs.String(), want)
		}
	}
}
