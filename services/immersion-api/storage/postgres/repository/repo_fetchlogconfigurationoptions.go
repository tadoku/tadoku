package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FetchLogConfigurationOptions(ctx context.Context) (*query.FetchLogConfigurationOptionsResponse, error) {
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

	options := query.FetchLogConfigurationOptionsResponse{
		Languages:  make([]query.Language, len(langs)),
		Activities: make([]query.Activity, len(acts)),
		Units:      make([]query.Unit, len(units)),
		Tags:       make([]query.Tag, len(tags)),
	}

	for i, l := range langs {
		options.Languages[i] = query.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	for i, a := range acts {
		options.Activities[i] = query.Activity{
			ID:      a.ID,
			Name:    a.Name,
			Default: a.Default,
		}
	}

	for i, u := range units {
		options.Units[i] = query.Unit{
			ID:            u.ID,
			LogActivityID: int(u.LogActivityID),
			Name:          u.Name,
			Modifier:      u.Modifier,
			LanguageCode:  postgres.NewStringFromNullString(u.LanguageCode),
		}
	}

	for i, t := range tags {
		options.Tags[i] = query.Tag{
			ID:            t.ID,
			LogActivityID: int(t.LogActivityID),
			Name:          t.Name,
		}
	}

	return &options, err
}
