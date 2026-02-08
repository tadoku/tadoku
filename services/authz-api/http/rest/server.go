package rest

import (
	"github.com/tadoku/tadoku/services/authz-api/domain"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi/internalapi"
)

type Server struct {
	roleGet                *domain.RoleGet
	roleUpdate             *domain.RoleUpdate
	publicPermissionCheck  *domain.PublicPermissionCheck
	internalPermissionCheck *domain.InternalPermissionCheck
	relationshipWriter     *domain.RelationshipWriter
}

func NewServer(
	roleGet *domain.RoleGet,
	roleUpdate *domain.RoleUpdate,
	publicPermissionCheck *domain.PublicPermissionCheck,
	internalPermissionCheck *domain.InternalPermissionCheck,
	relationshipWriter *domain.RelationshipWriter,
) *Server {
	return &Server{
		roleGet:                roleGet,
		roleUpdate:             roleUpdate,
		publicPermissionCheck:  publicPermissionCheck,
		internalPermissionCheck: internalPermissionCheck,
		relationshipWriter:     relationshipWriter,
	}
}

var _ openapi.ServerInterface = (*Server)(nil)
var _ internalapi.ServerInterface = (*Server)(nil)

