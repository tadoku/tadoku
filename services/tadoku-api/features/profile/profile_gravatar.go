package profile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

const maxAvatarBytes = 256 << 10

var allowedAvatarTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type Gravatar struct {
	baseURL string
	client  *http.Client
}

// NewGravatar takes a base URL without a trailing slash. The client's timeout bounds each fetch.
func NewGravatar(baseURL string, client *http.Client) *Gravatar {
	return &Gravatar{baseURL: baseURL, client: client}
}

// fetch returns a nil avatar when Gravatar has no image for the email.
func (g *Gravatar) fetch(ctx context.Context, email identity.Email) (*Avatar, error) {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email.String()))))
	url := g.baseURL + "/avatar/" + hex.EncodeToString(sum[:]) + "?s=128&d=404"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build gravatar request: %w", err)
	}

	response, err := g.client.Do(request)
	if err != nil {
		return nil, errx.NewUnavailableError("fetch gravatar", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, errx.NewUnavailableError(fmt.Sprintf("gravatar answered %d", response.StatusCode), nil)
	}

	contentType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || !allowedAvatarTypes[contentType] {
		return nil, errx.NewUnavailableError("gravatar answered an unsupported content type", err)
	}

	image, err := io.ReadAll(io.LimitReader(response.Body, maxAvatarBytes+1))
	if err != nil {
		return nil, errx.NewUnavailableError("read gravatar image", err)
	}
	if len(image) > maxAvatarBytes {
		return nil, errx.NewUnavailableError("gravatar image exceeds 256 KiB", nil)
	}
	return &Avatar{ContentType: contentType, Image: image}, nil
}
