package posts

import "github.com/jackc/pgx/v5/pgxpool"

type PostsRepository struct {
	db *pgxpool.Pool
}

func NewPostsRepository(db *pgxpool.Pool) *PostsRepository {
	return &PostsRepository{
		db: db,
	}
}
