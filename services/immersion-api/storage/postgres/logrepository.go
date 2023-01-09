package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/logcommand"
	"github.com/tadoku/tadoku/services/immersion-api/domain/logquery"
)

type LogRepository struct {
	psql *sql.DB
	q    *Queries
}

func NewLogRepository(psql *sql.DB) *LogRepository {
	return &LogRepository{
		psql: psql,
		q:    &Queries{psql},
	}
}

// COMMANDS
func (r *LogRepository) CreateLog(context.Context, *logcommand.LogCreateRequest) error {
	return nil
}

// QUERIES

func (r *LogRepository) FetchLogConfigurationOptions(ctx context.Context) (*logquery.FetchLogConfigurationOptionsResponse, error) {
	langs, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	acts, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	units, err := r.q.ListUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	tags, err := r.q.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	options := logquery.FetchLogConfigurationOptionsResponse{
		Languages:  make([]logquery.Language, len(langs)),
		Activities: make([]logquery.Activity, len(acts)),
		Units:      make([]logquery.Unit, len(units)),
		Tags:       make([]logquery.Tag, len(tags)),
	}

	for i, l := range langs {
		options.Languages[i] = logquery.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	for i, a := range acts {
		options.Activities[i] = logquery.Activity{
			ID:      a.ID,
			Name:    a.Name,
			Default: a.Default,
		}
	}

	for i, u := range units {
		options.Units[i] = logquery.Unit{
			ID:            u.ID,
			LogActivityID: int(u.LogActivityID),
			Name:          u.Name,
			Modifier:      u.Modifier,
			LanguageCode:  NewStringFromNullString(u.LanguageCode),
		}
	}

	for i, t := range tags {
		options.Tags[i] = logquery.Tag{
			ID:            t.ID,
			LogActivityID: int(t.LogActivityID),
			Name:          t.Name,
		}
	}

	return &options, err
}
