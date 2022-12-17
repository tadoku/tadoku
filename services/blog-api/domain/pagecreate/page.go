package pagecreate

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrPageAlreadyExists = errors.New("page with given id already exists")
var ErrInvalidPage = errors.New("unable to validate page")

type PageCreateRequest struct {
	ID          uuid.UUID  `json:"id" validate:"required"`
	Slug        string     `json:"slug" validate:"required,gt=1,lowercase"`
	Title       string     `json:"title" validate:"required"`
	Html        string     `json:"html" validate:"required"`
	PublishedAt *time.Time `json:"published_at`
}

type PageCreateResponse struct {
	ID    uuid.UUID `json:"id"`
	Slug  string    `json:"slug"`
	Title string    `json:"title"`
	Html  string    `json:"html"`
}
