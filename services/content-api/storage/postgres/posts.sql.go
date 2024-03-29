// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: posts.sql

package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createPost = `-- name: CreatePost :one
insert into posts (
  id,
  "namespace",
  slug,
  current_content_id,
  published_at
) values (
  $1,
  $2,
  $3,
  $4,
  $5
) returning id
`

type CreatePostParams struct {
	ID               uuid.UUID
	Namespace        string
	Slug             string
	CurrentContentID uuid.UUID
	PublishedAt      sql.NullTime
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, createPost,
		arg.ID,
		arg.Namespace,
		arg.Slug,
		arg.CurrentContentID,
		arg.PublishedAt,
	)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const createPostContent = `-- name: CreatePostContent :one
insert into posts_content (
  id,
  post_id,
  title,
  content
) values (
  $1,
  $2,
  $3,
  $4
) returning id
`

type CreatePostContentParams struct {
	ID      uuid.UUID
	PostID  uuid.UUID
	Title   string
	Content string
}

func (q *Queries) CreatePostContent(ctx context.Context, arg CreatePostContentParams) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, createPostContent,
		arg.ID,
		arg.PostID,
		arg.Title,
		arg.Content,
	)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const findPostBySlug = `-- name: FindPostBySlug :one
select
  posts.id,
  "namespace",
  slug,
  posts_content.title,
  posts_content.content,
  published_at
from posts
inner join posts_content
  on posts_content.id = posts.current_content_id
where
  deleted_at is null
  and "namespace" = $1
  and slug = $2
`

type FindPostBySlugParams struct {
	Namespace string
	Slug      string
}

type FindPostBySlugRow struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	Content     string
	PublishedAt sql.NullTime
}

func (q *Queries) FindPostBySlug(ctx context.Context, arg FindPostBySlugParams) (FindPostBySlugRow, error) {
	row := q.db.QueryRowContext(ctx, findPostBySlug, arg.Namespace, arg.Slug)
	var i FindPostBySlugRow
	err := row.Scan(
		&i.ID,
		&i.Namespace,
		&i.Slug,
		&i.Title,
		&i.Content,
		&i.PublishedAt,
	)
	return i, err
}

const listPosts = `-- name: ListPosts :many
select
  posts.id,
  "namespace",
  slug,
  posts_content.title,
  posts_content.content,
  published_at,
  posts.created_at,
  posts.updated_at
from posts
inner join posts_content
  on posts_content.id = posts.current_content_id
where
  deleted_at is null
  and ($1::boolean or published_at is not null)
  and "namespace" = $2
order by posts.created_at desc
limit $4
offset $3
`

type ListPostsParams struct {
	IncludeDrafts bool
	Namespace     string
	StartFrom     int32
	PageSize      int32
}

type ListPostsRow struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	Content     string
	PublishedAt sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (q *Queries) ListPosts(ctx context.Context, arg ListPostsParams) ([]ListPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, listPosts,
		arg.IncludeDrafts,
		arg.Namespace,
		arg.StartFrom,
		arg.PageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListPostsRow
	for rows.Next() {
		var i ListPostsRow
		if err := rows.Scan(
			&i.ID,
			&i.Namespace,
			&i.Slug,
			&i.Title,
			&i.Content,
			&i.PublishedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const postsMetadata = `-- name: PostsMetadata :one
select
  count(posts.id) as total_size,
  $1::boolean as drafts_included
from posts
where
  deleted_at is null
  and ($1::boolean or published_at is not null)
  and "namespace" = $2
`

type PostsMetadataParams struct {
	IncludeDrafts bool
	Namespace     string
}

type PostsMetadataRow struct {
	TotalSize      int64
	DraftsIncluded bool
}

func (q *Queries) PostsMetadata(ctx context.Context, arg PostsMetadataParams) (PostsMetadataRow, error) {
	row := q.db.QueryRowContext(ctx, postsMetadata, arg.IncludeDrafts, arg.Namespace)
	var i PostsMetadataRow
	err := row.Scan(&i.TotalSize, &i.DraftsIncluded)
	return i, err
}

const updatePost = `-- name: UpdatePost :one
update posts
set
  slug = $1,
  current_content_id = $2,
  published_at = $3,
  updated_at = now()
where
  id = $4 and
  deleted_at is null
returning id
`

type UpdatePostParams struct {
	Slug             string
	CurrentContentID uuid.UUID
	PublishedAt      sql.NullTime
	ID               uuid.UUID
}

func (q *Queries) UpdatePost(ctx context.Context, arg UpdatePostParams) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, updatePost,
		arg.Slug,
		arg.CurrentContentID,
		arg.PublishedAt,
		arg.ID,
	)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}
