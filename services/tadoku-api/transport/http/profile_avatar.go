package http

import (
	"bytes"
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionProfileAvatarByUserID(
	ctx context.Context,
	request openapi.ImmersionProfileAvatarByUserIDRequestObject,
) (openapi.ImmersionProfileAvatarByUserIDResponseObject, error) {
	avatar, err := s.application.FindAvatar(ctx, request.UserId)
	if err != nil {
		s.logOperationError(ctx, "FindAvatar", err)
		return nil, err
	}

	return openapi.ImmersionProfileAvatarByUserID200ImageResponse{
		Body: bytes.NewReader(avatar.Image),
		Headers: openapi.ImmersionProfileAvatarByUserID200ResponseHeaders{
			CacheControl:        "public, max-age=86400",
			XContentTypeOptions: "nosniff",
		},
		ContentType:   avatar.ContentType,
		ContentLength: int64(len(avatar.Image)),
	}, nil
}
