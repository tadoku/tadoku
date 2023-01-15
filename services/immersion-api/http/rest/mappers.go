package rest

import (
	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestquery"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

func contestRegistrationToAPI(r *contestquery.ContestRegistration) *openapi.ContestRegistration {

	registration := openapi.ContestRegistration{
		ContestId:       r.ContestID,
		Id:              &r.ID,
		Languages:       make([]openapi.Language, len(r.Languages)),
		UserId:          r.UserID,
		UserDisplayName: r.UserDisplayName,
	}

	if r.Contest != nil {
		contest := openapi.ContestView{
			Id:                &r.ContestID,
			ContestStart:      types.Date{Time: r.Contest.ContestStart},
			ContestEnd:        types.Date{Time: r.Contest.ContestEnd},
			RegistrationEnd:   types.Date{Time: r.Contest.RegistrationEnd},
			Title:             r.Contest.Title,
			Description:       r.Contest.Description,
			Official:          r.Contest.Official,
			Private:           r.Contest.Private,
			AllowedLanguages:  []openapi.Language{},
			AllowedActivities: make([]openapi.Activity, len(r.Contest.AllowedActivities)),
		}

		for i, a := range r.Contest.AllowedActivities {
			contest.AllowedActivities[i] = openapi.Activity{
				Id:   a.ID,
				Name: a.Name,
			}
		}

		registration.Contest = &contest
	}

	for i, lang := range r.Languages {
		registration.Languages[i] = openapi.Language{
			Code: lang.Code,
			Name: lang.Name,
		}
	}

	return &registration
}
