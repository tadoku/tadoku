package kratosidentity

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	kratosapi "github.com/ory/kratos-client-go"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Outcome uint8

const (
	Applied Outcome = iota + 1
	SkippedForTestTenant
)

// Pass a context carrying an explicit tenant to every write operation.
type Writer struct {
	client *kratosapi.APIClient
	logger *slog.Logger
}

// Supply an SDK client whose HTTP transport has bounded request timeouts.
func NewWriter(client *kratosapi.APIClient, logger *slog.Logger) *Writer {
	return &Writer{client: client, logger: logger}
}

func (writer *Writer) Deactivate(ctx context.Context, id uuid.UUID) (Outcome, error) {
	if outcome, err := writer.guard(ctx, id); outcome != Applied {
		return outcome, err
	}

	statePatch := kratosapi.NewJsonPatch("replace", "/state")
	statePatch.SetValue(kratosapi.IDENTITYSTATE_INACTIVE)
	req := writer.client.IdentityApi.PatchIdentity(ctx, id.String()).JsonPatch([]kratosapi.JsonPatch{*statePatch})
	_, response, err := writer.client.IdentityApi.PatchIdentityExecute(req)
	return writeOutcome(response, err, "deactivate identity")
}

func (writer *Writer) DeleteSessions(ctx context.Context, id uuid.UUID) (Outcome, error) {
	if outcome, err := writer.guard(ctx, id); outcome != Applied {
		return outcome, err
	}

	req := writer.client.IdentityApi.DeleteIdentitySessions(ctx, id.String())
	response, err := writer.client.IdentityApi.DeleteIdentitySessionsExecute(req)
	return writeOutcome(response, err, "delete identity sessions")
}

func (writer *Writer) Delete(ctx context.Context, id uuid.UUID) (Outcome, error) {
	if outcome, err := writer.guard(ctx, id); outcome != Applied {
		return outcome, err
	}

	req := writer.client.IdentityApi.DeleteIdentity(ctx, id.String())
	response, err := writer.client.IdentityApi.DeleteIdentityExecute(req)
	return writeOutcome(response, err, "delete identity")
}

func (writer *Writer) guard(ctx context.Context, id uuid.UUID) (Outcome, error) {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return 0, fmt.Errorf("kratos identity write requires a tenant")
	}
	if key != tenant.Production() {
		writer.logger.InfoContext(
			ctx,
			"skipped Kratos identity write for test tenant",
			"tenant", key.String(),
			"identity_id", id,
		)
		return SkippedForTestTenant, nil
	}
	return Applied, nil
}

func writeOutcome(response *http.Response, err error, operation string) (Outcome, error) {
	if err != nil && (response == nil || response.StatusCode != http.StatusNotFound) {
		return 0, fmt.Errorf("could not %s: %w", operation, err)
	}
	return Applied, nil
}
