package http

import (
	"context"
	"io"
	stdhttp "net/http"
	"strings"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func withAuthzPermissionCheckEmptyBodyCompatibility(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, request *stdhttp.Request) {
		// Echo accepts an empty body and lets the domain return guest 401 or authenticated 400;
		// the generated required-body decoder rejects EOF before the domain runs.
		if request.Method == stdhttp.MethodPost && request.URL.Path == "/authz/permission/check" && request.ContentLength == 0 {
			request.Body = io.NopCloser(strings.NewReader("null"))
			request.ContentLength = 4
		}
		next.ServeHTTP(w, request)
	})
}

func (s *server) AuthzPermissionCheck(
	ctx context.Context,
	request openapi.AuthzPermissionCheckRequestObject,
) (openapi.AuthzPermissionCheckResponseObject, error) {
	parameters := app.PermissionCheckParameters{}
	if request.Body != nil {
		parameters = app.PermissionCheckParameters{
			Namespace: request.Body.Namespace,
			Object:    request.Body.Object,
			Relation:  request.Body.Relation,
		}
	}

	err := s.application.CheckPermission(ctx, parameters)
	if err != nil {
		s.logOperationError(ctx, "check permission", err)
		return nil, err
	}

	return openapi.AuthzPermissionCheck200JSONResponse{Allowed: false}, nil
}
