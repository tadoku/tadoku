package rest

import (
	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

func domainLeaderboardToAPI(leaderboard domain.Leaderboard) *openapi.Leaderboard {
	res := openapi.Leaderboard{
		Entries:       make([]openapi.LeaderboardEntry, len(leaderboard.Entries)),
		NextPageToken: leaderboard.NextPageToken,
		TotalSize:     leaderboard.TotalSize,
	}

	for i, entry := range leaderboard.Entries {
		res.Entries[i] = openapi.LeaderboardEntry{
			Rank:            entry.Rank,
			UserId:          entry.UserID,
			UserDisplayName: entry.UserDisplayName,
			Score:           entry.Score,
			IsTie:           entry.IsTie,
		}
	}

	return &res
}

func logToAPI(log *domain.Log) *openapi.Log {
	refs := make([]openapi.ContestRegistrationReference, len(log.Registrations))
	for i, it := range log.Registrations {
		refs[i] = openapi.ContestRegistrationReference{
			ContestId:            it.ContestID,
			ContestEnd:           types.Date{Time: it.ContestEnd},
			RegistrationId:       it.RegistrationID,
			Title:                it.Title,
			OwnerUserDisplayName: &it.OwnerUserDisplayName,
			Official:             &it.Official,
			Score:                &it.Score,
		}
	}

	return &openapi.Log{
		Id: log.ID,
		Activity: openapi.Activity{
			Id:        int32(log.ActivityID),
			Name:      log.ActivityName,
			InputType: openapi.ActivityInputType(log.ActivityInputType),
		},
		Language: openapi.Language{
			Code: log.LanguageCode,
			Name: log.LanguageName,
		},
		Amount:          log.Amount,
		Modifier:        log.Modifier,
		Score:           log.EffectiveScore(),
		Tags:            log.Tags,
		UnitId:          log.UnitID,
		UnitName:        log.UnitName,
		DurationSeconds: intPtrFromInt32Ptr(log.DurationSeconds),
		UserId:          log.UserID,
		UserDisplayName: log.UserDisplayName,
		CreatedAt:       log.CreatedAt,
		Deleted:         log.Deleted,
		Description:     log.Description,
		Registrations:   &refs,
	}
}

func intPtrFromInt32Ptr(v *int32) *int {
	if v == nil {
		return nil
	}
	i := int(*v)
	return &i
}

func int32PtrFromIntPtr(v *int) *int32 {
	if v == nil {
		return nil
	}
	i := int32(*v)
	return &i
}

func contestRegistrationToAPI(r *domain.ContestRegistration) *openapi.ContestRegistration {
	registration := openapi.ContestRegistration{
		ContestId:       r.ContestID,
		Id:              &r.ID,
		Languages:       make([]openapi.Language, len(r.Languages)),
		UserId:          r.UserID,
		UserDisplayName: r.UserDisplayName,
	}

	if r.Contest != nil {
		contest := openapi.ContestView{
			Id:                   &r.ContestID,
			ContestStart:         types.Date{Time: r.Contest.ContestStart},
			ContestEnd:           types.Date{Time: r.Contest.ContestEnd},
			RegistrationEnd:      types.Date{Time: r.Contest.RegistrationEnd},
			Title:                r.Contest.Title,
			Description:          r.Contest.Description,
			Official:             r.Contest.Official,
			Private:              r.Contest.Private,
			OwnerUserId:          &r.Contest.OwnerUserID,
			OwnerUserDisplayName: &r.Contest.OwnerUserDisplayName,
			AllowedLanguages:     []openapi.Language{},
			AllowedActivities:    make([]openapi.Activity, len(r.Contest.AllowedActivities)),
		}

		for i, a := range r.Contest.AllowedActivities {
			contest.AllowedActivities[i] = openapi.Activity{
				Id:        a.ID,
				Name:      a.Name,
				InputType: openapi.ActivityInputType(a.InputType),
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
