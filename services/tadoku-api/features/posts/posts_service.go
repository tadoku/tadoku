package posts

import (
	"context"
	"math"
	"strconv"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Service struct {
	posts *PostsRepository
}

func NewService(posts *PostsRepository) *Service {
	return &Service{posts: posts}
}

func (s *Service) CreatePost(ctx context.Context, parameters CreatePostParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	now := timex.Now()
	item := &Post{
		ID:          parameters.ID,
		Namespace:   parameters.Namespace,
		Slug:        parameters.Slug,
		Title:       parameters.Title,
		Content:     parameters.Content,
		PublishedAt: parameters.PublishedAt,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}
	return s.posts.CreatePost(ctx, item)
}

func (s *Service) DeletePost(ctx context.Context, namespace string, id uuid.UUID) error {
	return s.posts.DeletePost(ctx, namespace, id, timex.Now())
}

func (s *Service) FindPostBySlug(ctx context.Context, namespace, slug string) (*Post, error) {
	if namespace == "" {
		return nil, ErrInvalidNamespace
	}
	if slug == "" {
		return nil, errx.NewInvalidInputError("slug is required")
	}

	post, err := s.posts.FindPostBySlug(ctx, namespace, slug)
	if err != nil {
		return nil, err
	}
	if post.PublishedAt == nil || post.PublishedAt.After(timex.Now()) {
		return nil, ErrPostNotFound
	}

	return post, nil
}

func (s *Service) FindPostByID(ctx context.Context, namespace string, id uuid.UUID) (*Post, error) {
	return s.posts.FindPostByID(ctx, namespace, id)
}

func (s *Service) ListPosts(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*PostList, error) {
	if namespace == "" {
		return nil, ErrInvalidNamespace
	}
	if pageSize < 0 || page < 0 {
		return nil, ErrInvalidPagination
	}
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := int64(page)
	if page > math.MaxInt64/pageSize {
		offset = math.MaxInt64
	} else {
		offset *= int64(pageSize)
	}

	posts, totalSize, err := s.posts.ListPosts(ctx, namespace, includeDrafts, timex.Now(), int32(pageSize), offset)
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if int64(pageSize) < int64(totalSize)-offset {
		nextPageToken = strconv.Itoa(page + 1)
	}

	return &PostList{
		Posts:         posts,
		TotalSize:     totalSize,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Service) UpdatePost(ctx context.Context, parameters UpdatePostParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	post, err := s.posts.FindPostByID(ctx, parameters.Namespace, parameters.ID)
	if err != nil {
		return err
	}

	contentChanged := post.Title != parameters.Title || post.Content != parameters.Content
	post.Slug = parameters.Slug
	post.Title = parameters.Title
	post.Content = parameters.Content
	post.PublishedAt = parameters.PublishedAt
	now := timex.Now()
	post.UpdatedAt = &now

	return s.posts.UpdatePost(ctx, post, contentChanged)
}

func (s *Service) GetPostVersion(ctx context.Context, namespace string, postID, contentID uuid.UUID) (*PostVersion, error) {
	return s.posts.GetPostVersion(ctx, namespace, postID, contentID)
}

func (s *Service) ListPostVersions(ctx context.Context, namespace string, id uuid.UUID) ([]PostVersion, error) {
	return s.posts.ListPostVersions(ctx, namespace, id)
}
