package http

import (
	"log/slog"
	stdhttp "net/http"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func listActiveAnnouncements(application *app.Application, logger *slog.Logger) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		items, err := application.ListActiveAnnouncements(r.Context(), r.PathValue("namespace"))
		if err != nil {
			logger.ErrorContext(r.Context(), "list active announcements failed",
				"correlation_id", correlationID(r),
				"error", err,
			)
			w.WriteHeader(stdhttp.StatusInternalServerError)
			return
		}

		response := openapi.ContentAnnouncements{
			Announcements: make([]openapi.ContentAnnouncement, 0, len(items)),
		}
		for _, item := range items {
			response.Announcements = append(response.Announcements, openapi.ContentAnnouncement{
				Id:        &item.ID,
				Namespace: &item.Namespace,
				Title:     item.Title,
				Content:   item.Content,
				Style:     openapi.ContentAnnouncementStyle(item.Style),
				Href:      item.Href,
				StartsAt:  item.StartsAt,
				EndsAt:    item.EndsAt,
				CreatedAt: &item.CreatedAt,
				UpdatedAt: &item.UpdatedAt,
			})
		}

		writeJSON(w, stdhttp.StatusOK, response)
	})
}
