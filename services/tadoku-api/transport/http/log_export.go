package http

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/oapi-codegen/nullable"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (s *server) ImmersionLogExport(
	ctx context.Context,
	request openapi.ImmersionLogExportRequestObject,
) (openapi.ImmersionLogExportResponseObject, error) {
	if !request.Params.Format.Valid() {
		return nil, errx.NewInvalidInputError("format must be csv or json")
	}

	logs, err := s.application.ExportOwnLogs(ctx)
	if err != nil {
		s.logOperationError(ctx, "export own logs", err)
		return nil, err
	}

	disposition := `attachment; filename="tadoku-logs.` + string(request.Params.Format) + `"`
	headers := openapi.ImmersionLogExport200ResponseHeaders{ContentDisposition: &disposition}
	if request.Params.Format == openapi.Json {
		body := logExportJSON(logs)
		return openapi.ImmersionLogExport200JSONResponse{Body: body, Headers: headers}, nil
	}

	body, err := logExportCSV(logs)
	if err != nil {
		return nil, err
	}

	return openapi.ImmersionLogExport200TextcsvResponse{
		Body:          body,
		Headers:       headers,
		ContentLength: int64(body.Len()),
	}, nil
}

func logExportJSON(logs []app.ExportedLog) openapi.ImmersionLogExport {
	body := openapi.ImmersionLogExport{Logs: make([]openapi.ImmersionLogExportEntry, 0, len(logs))}

	for _, log := range logs {
		entry := openapi.ImmersionLogExportEntry{
			Id:              log.ID,
			CreatedAt:       log.CreatedAt.UTC(),
			LanguageCode:    log.LanguageCode,
			LanguageName:    log.LanguageName,
			ActivityId:      log.Activity.ID,
			ActivityName:    log.Activity.Name,
			Amount:          nullable.NewNullNullable[float32](),
			UnitName:        nullable.NewNullNullable[string](),
			DurationSeconds: nullable.NewNullNullable[int32](),
			Score:           log.Score,
			Tags:            log.Tags,
			Description:     nullable.NewNullNullable[string](),
		}

		if log.Amount != nil {
			entry.Amount.Set(log.Amount.Value)
			entry.UnitName.Set(log.Amount.UnitName)
		}
		if log.DurationSeconds != nil {
			entry.DurationSeconds.Set(*log.DurationSeconds)
		}
		if log.Description != nil {
			entry.Description.Set(*log.Description)
		}

		body.Logs = append(body.Logs, entry)
	}

	return body
}

func logExportCSV(logs []app.ExportedLog) (*bytes.Buffer, error) {
	var body bytes.Buffer
	body.WriteString("\xef\xbb\xbf")
	writer := csv.NewWriter(&body)
	if err := writer.Write([]string{
		"id",
		"created_at",
		"language_code",
		"language_name",
		"activity_id",
		"activity_name",
		"amount",
		"unit_name",
		"duration_seconds",
		"score",
		"tags",
		"description",
	}); err != nil {
		return nil, err
	}

	for _, log := range logs {
		var amount, unitName, duration, description string
		if log.Amount != nil {
			amount = strconv.FormatFloat(float64(log.Amount.Value), 'f', -1, 32)
			unitName = log.Amount.UnitName
		}
		if log.DurationSeconds != nil {
			duration = strconv.FormatInt(int64(*log.DurationSeconds), 10)
		}
		if log.Description != nil {
			description = *log.Description
		}

		if err := writer.Write([]string{
			log.ID.String(),
			log.CreatedAt.UTC().Format(time.RFC3339Nano),
			guardCSVText(log.LanguageCode),
			guardCSVText(log.LanguageName),
			strconv.FormatInt(int64(log.Activity.ID), 10),
			guardCSVText(log.Activity.Name),
			amount,
			guardCSVText(unitName),
			duration,
			strconv.FormatFloat(float64(log.Score), 'f', -1, 32),
			guardCSVText(strings.Join(log.Tags, ", ")),
			guardCSVText(description),
		}); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return &body, nil
}

func guardCSVText(value string) string {
	if value != "" && strings.ContainsRune("=+-@\t\r", rune(value[0])) {
		return "'" + value
	}
	return value
}
