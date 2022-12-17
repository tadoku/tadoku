package pagecreate

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrPageAlreadyExists = errors.New("page with given id already exists")
var ErrInvalidPage = errors.New("unable to validate page")

type PageCreateRequest struct {
	ID          uuid.UUID `validate:"required"`
	Slug        string    `validate:"required,gt=1,lowercase"`
	Title       string    `validate:"required"`
	Html        string    `validate:"required"`
	PublishedAt *time.Time
}

type PageCreateResponse struct {
	ID    uuid.UUID
	Slug  string
	Title string
	Html  string
}
