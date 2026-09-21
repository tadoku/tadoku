package http

import (
	"context"
	"io"
	stdhttp "net/http"
	"strings"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func withAuthzRoleUpdateEmptyBodyCompatibility(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, request *stdhttp.Request) {
		// Echo accepts an empty body and lets the domain authorize before returning invalid input;
		// the generated required-body decoder rejects EOF before the domain runs.
		if request.Method == stdhttp.MethodPut && strings.HasPrefix(request.URL.Path, "/authz/users/") && strings.HasSuffix(request.URL.Path, "/role") && request.ContentLength == 0 {
			request.Body = io.NopCloser(strings.NewReader("null"))
			request.ContentLength = 4
		}
		next.ServeHTTP(w, request)
	})
}

func (s *server) AuthzRoleUpdate(
	ctx context.Context,
	request openapi.AuthzRoleUpdateRequestObject,
) (openapi.AuthzRoleUpdateResponseObject, error) {
	parameters := app.RoleUpdateParameters{UserID: request.Id}
	if request.Body != nil {
		parameters.Role = app.Role(request.Body.Role)
		parameters.Reason = request.Body.Reason
	}

	if err := s.application.UpdateRole(ctx, parameters); err != nil {
		s.logOperationError(ctx, "update user role", err)
		return nil, err
	}

	return openapi.AuthzRoleUpdate200Response{}, nil
}
