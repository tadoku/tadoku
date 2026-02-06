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
	pageFindByID *domain.PageFindByID,
	pageList *domain.PageList,
	pageVersionList *domain.PageVersionList,
	pageVersionGet *domain.PageVersionGet,
	postCreate *domain.PostCreate,
	postUpdate *domain.PostUpdate,
	postDelete *domain.PostDelete,
	postFind *domain.PostFind,
	postFindByID *domain.PostFindByID,
	postList *domain.PostList,
	postVersionList *domain.PostVersionList,
	postVersionGet *domain.PostVersionGet,
) openapi.ServerInterface {
	return &Server{
		pageCreate:      pageCreate,
		pageUpdate:      pageUpdate,
		pageDelete:      pageDelete,
		pageFind:        pageFind,
		pageFindByID:    pageFindByID,
		pageList:        pageList,
		pageVersionList: pageVersionList,
		pageVersionGet:  pageVersionGet,
		postCreate:      postCreate,
		postUpdate:      postUpdate,
		postDelete:      postDelete,
		postFind:        postFind,
		postFindByID:    postFindByID,
		postList:        postList,
		postVersionList: postVersionList,
		postVersionGet:  postVersionGet,
	}
}

type Server struct {
	pageCreate      *domain.PageCreate
	pageUpdate      *domain.PageUpdate
	pageDelete      *domain.PageDelete
	pageFind        *domain.PageFind
	pageFindByID    *domain.PageFindByID
	pageList        *domain.PageList
	pageVersionList *domain.PageVersionList
	pageVersionGet  *domain.PageVersionGet

	postCreate      *domain.PostCreate
	postUpdate      *domain.PostUpdate
	postDelete      *domain.PostDelete
	postFind        *domain.PostFind
	postFindByID    *domain.PostFindByID
	postList        *domain.PostList
	postVersionList *domain.PostVersionList
	postVersionGet  *domain.PostVersionGet
}
