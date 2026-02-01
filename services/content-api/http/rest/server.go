package rest

import (
	"github.com/tadoku/tadoku/services/content-api/domain"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	pageCreate *domain.PageCreate,
	pageUpdate *domain.PageUpdate,
	pageFind *domain.PageFind,
	pageList *domain.PageList,
	postCreate *domain.PostCreate,
	postUpdate *domain.PostUpdate,
	postFind *domain.PostFind,
	postList *domain.PostList,
) openapi.ServerInterface {
	return &Server{
		pageCreate: pageCreate,
		pageUpdate: pageUpdate,
		pageFind:   pageFind,
		pageList:   pageList,
		postCreate: postCreate,
		postUpdate: postUpdate,
		postFind:   postFind,
		postList:   postList,
	}
}

type Server struct {
	pageCreate *domain.PageCreate
	pageUpdate *domain.PageUpdate
	pageFind   *domain.PageFind
	pageList   *domain.PageList

	postCreate *domain.PostCreate
	postUpdate *domain.PostUpdate
	postFind   *domain.PostFind
	postList   *domain.PostList
}
