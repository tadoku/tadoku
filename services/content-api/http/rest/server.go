package rest

import (
	"github.com/tadoku/tadoku/services/content-api/domain"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	pageCreate *domain.PageCreate,
	pageUpdate *domain.PageUpdate,
	pageDelete *domain.PageDelete,
	pageFind *domain.PageFind,
	pageList *domain.PageList,
	pageVersionList *domain.PageVersionList,
	postCreate *domain.PostCreate,
	postUpdate *domain.PostUpdate,
	postDelete *domain.PostDelete,
	postFind *domain.PostFind,
	postList *domain.PostList,
	postVersionList *domain.PostVersionList,
) openapi.ServerInterface {
	return &Server{
		pageCreate:      pageCreate,
		pageUpdate:      pageUpdate,
		pageDelete:      pageDelete,
		pageFind:        pageFind,
		pageList:        pageList,
		pageVersionList: pageVersionList,
		postCreate:      postCreate,
		postUpdate:      postUpdate,
		postDelete:      postDelete,
		postFind:        postFind,
		postList:        postList,
		postVersionList: postVersionList,
	}
}

type Server struct {
	pageCreate      *domain.PageCreate
	pageUpdate      *domain.PageUpdate
	pageDelete      *domain.PageDelete
	pageFind        *domain.PageFind
	pageList        *domain.PageList
	pageVersionList *domain.PageVersionList

	postCreate      *domain.PostCreate
	postUpdate      *domain.PostUpdate
	postDelete      *domain.PostDelete
	postFind        *domain.PostFind
	postList        *domain.PostList
	postVersionList *domain.PostVersionList
}
