package rest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/content-api/domain"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// COMMANDS

// Creates a new announcement
// (POST /announcements/{namespace})
func (s *Server) AnnouncementCreate(ctx echo.Context, namespace string) error {
	var req openapi.AnnouncementCreateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	id := uuid.New()
	if req.Id != nil {
		id = *req.Id
	}

	resp, err := s.announcementCreate.Execute(ctx.Request().Context(), &domain.AnnouncementCreateRequest{
		ID:        id,
		Namespace: namespace,
		Title:     req.Title,
		Content:   req.Content,
		Style:     string(req.Style),
		Href:      req.Href,
		StartsAt:  req.StartsAt,
		EndsAt:    req.EndsAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrInvalidAnnouncement) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusCreated, announcementToOpenAPI(resp.Announcement))
}

// Updates an existing announcement
// (PUT /announcements/{namespace}/{id})
func (s *Server) AnnouncementUpdate(ctx echo.Context, namespace string, id string) error {
	var req openapi.AnnouncementUpdateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	resp, err := s.announcementUpdate.Execute(ctx.Request().Context(), parsedID, &domain.AnnouncementUpdateRequest{
		Namespace: namespace,
		Title:     req.Title,
		Content:   req.Content,
		Style:     string(req.Style),
		Href:      req.Href,
		StartsAt:  req.StartsAt,
		EndsAt:    req.EndsAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrInvalidAnnouncement) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}
		if errors.Is(err, domain.ErrAnnouncementNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, announcementToOpenAPI(resp.Announcement))
}

// Deletes an existing announcement
// (DELETE /announcements/{namespace}/{id})
func (s *Server) AnnouncementDelete(ctx echo.Context, namespace string, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	err = s.announcementDelete.Execute(ctx.Request().Context(), parsedID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrAnnouncementNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// QUERIES

// Gets an announcement by ID
// (GET /announcements/{namespace}/{id})
func (s *Server) AnnouncementFindByID(ctx echo.Context, namespace string, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	announcement, err := s.announcementFindByID.Execute(ctx.Request().Context(), parsedID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrAnnouncementNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, announcementToOpenAPI(announcement))
}

// Lists all announcements (admin)
// (GET /announcements/{namespace})
func (s *Server) AnnouncementList(ctx echo.Context, namespace string, params openapi.AnnouncementListParams) error {
	pageSize := 0
	page := 0

	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	if params.Page != nil {
		page = *params.Page
	}

	resp, err := s.announcementList.Execute(ctx.Request().Context(), &domain.AnnouncementListRequest{
		Namespace: namespace,
		PageSize:  pageSize,
		Page:      page,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.AnnouncementList{
		Announcements: []openapi.Announcement{},
		NextPageToken: resp.NextPageToken,
		TotalSize:     resp.TotalSize,
	}

	for _, a := range resp.Announcements {
		res.Announcements = append(res.Announcements, announcementToOpenAPI(&a))
	}

	return ctx.JSON(http.StatusOK, res)
}

// Lists currently active announcements (public)
// (GET /announcements/{namespace}/active)
func (s *Server) AnnouncementListActive(ctx echo.Context, namespace string) error {
	resp, err := s.announcementListActive.Execute(ctx.Request().Context(), &domain.AnnouncementListActiveRequest{
		Namespace: namespace,
	})
	if err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.Announcements{
		Announcements: []openapi.Announcement{},
	}

	for _, a := range resp.Announcements {
		res.Announcements = append(res.Announcements, announcementToOpenAPI(&a))
	}

	return ctx.JSON(http.StatusOK, res)
}

func announcementToOpenAPI(a *domain.Announcement) openapi.Announcement {
	return openapi.Announcement{
		Id:        &a.ID,
		Namespace: &a.Namespace,
		Title:     a.Title,
		Content:   a.Content,
		Style:     openapi.AnnouncementStyle(a.Style),
		Href:      a.Href,
		StartsAt:  a.StartsAt,
		EndsAt:    a.EndsAt,
		CreatedAt: &a.CreatedAt,
		UpdatedAt: &a.UpdatedAt,
	}
}
