package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/authz-api/domain"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi/internalapi"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// (GET /internal/v1/ping)
func (s *Server) InternalPing(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}

// (POST /internal/v1/permission/check)
func (s *Server) InternalPermissionCheck(ctx echo.Context) error {
	var req internalapi.InternalPermissionCheckJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	subj, err := subjectFromInternal(req.SubjectId, req.SubjectSet)
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		return ctx.NoContent(http.StatusBadRequest)
	}

	allowed, err := s.internalPermissionCheck.Execute(ctx.Request().Context(), domain.InternalPermissionCheckRequest{
		Namespace: req.Namespace,
		Object:    req.Object,
		Relation:  req.Relation,
		Subject:   subj,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, internalapi.PermissionCheckResponse{Allowed: allowed})
}

// (POST /internal/v1/relationships)
func (s *Server) InternalRelationshipCreate(ctx echo.Context) error {
	var req internalapi.RelationshipWriteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	svc := commondomain.ParseServiceIdentity(ctx.Request().Context())

	subj, err := subjectFromInternal(req.SubjectId, req.SubjectSet)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	err = s.relationshipWriter.Create(ctx.Request().Context(), svc.Name, domain.RelationshipWriteRequest{
		Namespace: req.Namespace,
		Object:    req.Object,
		Relation:  req.Relation,
		Subject:   subj,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}

// (DELETE /internal/v1/relationships)
func (s *Server) InternalRelationshipDelete(ctx echo.Context) error {
	var req internalapi.RelationshipWriteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	svc := commondomain.ParseServiceIdentity(ctx.Request().Context())

	subj, err := subjectFromInternal(req.SubjectId, req.SubjectSet)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	err = s.relationshipWriter.Delete(ctx.Request().Context(), svc.Name, domain.RelationshipWriteRequest{
		Namespace: req.Namespace,
		Object:    req.Object,
		Relation:  req.Relation,
		Subject:   subj,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}

func subjectFromInternal(subjectID *string, subjectSet *internalapi.SubjectSet) (ketoclient.Subject, error) {
	switch {
	case subjectID != nil && *subjectID != "" && subjectSet != nil:
		return ketoclient.Subject{}, commondomain.ErrRequestInvalid
	case subjectID != nil && *subjectID != "":
		return ketoclient.Subject{ID: *subjectID}, nil
	case subjectSet != nil:
		return ketoclient.Subject{Set: &ketoclient.SubjectSet{
			Namespace: subjectSet.Namespace,
			Object:    subjectSet.Object,
			Relation:  subjectSet.Relation,
		}}, nil
	default:
		return ketoclient.Subject{}, commondomain.ErrRequestInvalid
	}
}
