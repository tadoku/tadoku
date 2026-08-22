package rest

import (
	"github.com/tadoku/tadoku/services/authz-api/domain"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi/internalapi"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi/proxyapi"
)

type Server struct {
	roleGet                 *domain.RoleGet
	roleUpdate              *domain.RoleUpdate
	publicPermissionCheck   *domain.PublicPermissionCheck
	internalPermissionCheck *domain.InternalPermissionCheck
	relationshipWriter      *domain.RelationshipWriter
	proxyAdminCheck         *domain.ProxyAdminCheck
}

func NewServer(
	roleGet *domain.RoleGet,
	roleUpdate *domain.RoleUpdate,
	publicPermissionCheck *domain.PublicPermissionCheck,
	internalPermissionCheck *domain.InternalPermissionCheck,
	relationshipWriter *domain.RelationshipWriter,
	proxyAdminCheck *domain.ProxyAdminCheck,
) *Server {
	return &Server{
		roleGet:                 roleGet,
		roleUpdate:              roleUpdate,
		publicPermissionCheck:   publicPermissionCheck,
		internalPermissionCheck: internalPermissionCheck,
		relationshipWriter:      relationshipWriter,
		proxyAdminCheck:         proxyAdminCheck,
	}
}

var _ openapi.ServerInterface = (*Server)(nil)
var _ internalapi.ServerInterface = (*Server)(nil)
var _ proxyapi.ServerInterface = (*Server)(nil)
