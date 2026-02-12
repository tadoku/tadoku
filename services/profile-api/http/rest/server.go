package rest

import (
	"github.com/tadoku/tadoku/services/profile-api/domain"
	"github.com/tadoku/tadoku/services/profile-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	userList *domain.UserList,
) openapi.ServerInterface {
	return &Server{
		userList: userList,
	}
}

type Server struct {
	userList *domain.UserList
}
