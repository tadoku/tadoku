package e2e_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	callbackopenapi "github.com/tadoku/tadoku/services/tadoku-api/generated/openapi/callback"
)

type accessLevel int

const (
	accessPublic accessLevel = iota
	accessAuthenticated
	accessAdmin
	accessCallback
)

// Every operation on the generated strict server interfaces must declare its access level.
var operationAccess = map[string]accessLevel{
	"ImmersionLogFindByID":     accessPublic,
	"ImmersionProfileListLogs": accessPublic,
	"ImmersionContestListLogs": accessPublic,

	"ImmersionContestProfileFetchScores":   accessPublic,
	"ImmersionContestProfileFetchActivity": accessPublic,
	"ImmersionContestFetchLeaderboard":     accessPublic,
	"ImmersionFetchLeaderboardForYear":     accessPublic,
	"ImmersionFetchLeaderboardGlobal":      accessPublic,

	"ImmersionProfileFindByUserID":                accessPublic,
	"ImmersionProfileYearlyActivityByUserID":      accessPublic,
	"ImmersionProfileYearlyScoresByUserID":        accessPublic,
	"ImmersionProfileYearlyActivitySplitByUserID": accessPublic,

	"AuthzPermissionCheck":      accessAuthenticated,
	"AuthzProxyProxyAdminCheck": accessCallback,
	"AuthzRoleGet":              accessPublic,
	"AuthzRoleUpdate":           accessAdmin,

	"ContentAnnouncementCreate":     accessAdmin,
	"ContentAnnouncementDelete":     accessAdmin,
	"ContentAnnouncementFindByID":   accessAdmin,
	"ContentAnnouncementList":       accessAdmin,
	"ContentAnnouncementListActive": accessPublic,
	"ContentAnnouncementUpdate":     accessAdmin,
	"ContentPageCreate":             accessAdmin,
	"ContentPageDelete":             accessAdmin,
	"ContentPageFindBySlug":         accessPublic,
	"ContentPageList":               accessAdmin,
	"ContentPageUpdate":             accessAdmin,
	"ContentPageVersionGet":         accessAdmin,
	"ContentPageVersionList":        accessAdmin,
	"ContentPostCreate":             accessAdmin,
	"ContentPostDelete":             accessAdmin,
	"ContentPostFindBySlug":         accessPublic,
	"ContentPostList":               accessPublic,
	"ContentPostUpdate":             accessAdmin,
	"ContentPostVersionGet":         accessAdmin,
	"ContentPostVersionList":        accessAdmin,

	"ImmersionContestCreate":                   accessAuthenticated,
	"ImmersionContestCreatePermissionCheck":    accessPublic,
	"ImmersionContestList":                     accessPublic,
	"ImmersionContestFindByID":                 accessPublic,
	"ImmersionContestFindOngoingRegistrations": accessAuthenticated,
	"ImmersionContestFindRegistration":         accessAuthenticated,
	"ImmersionContestFindLatestOfficial":       accessPublic,
	"ImmersionContestFetchSummary":             accessPublic,
	"ImmersionContestGetConfigurations":        accessPublic,
	"ImmersionContestRegistrationUpsert":       accessAuthenticated,

	"ImmersionProfileYearlyContestRegistrationsByUserID": accessPublic,
	"ImmersionLogGetConfigurations":                      accessAuthenticated,
	"ImmersionLogTagSuggestions":                         accessAuthenticated,
	"ImmersionScorePreview":                              accessAuthenticated,
	"ImmersionScoringRuleSetListPlatform":                accessAuthenticated,
	"ImmersionScoringRuleSetListContest":                 accessAuthenticated,
	"ImmersionScoringRuleSetCreatePlatform":              accessAdmin,
	"ImmersionScoringRuleSetCreateContest":               accessAuthenticated,

	"ImmersionLanguageList":         accessAdmin,
	"ImmersionLanguageCreate":       accessAdmin,
	"ImmersionLanguageUpdate":       accessAdmin,
	"ImmersionFeatureFlagDecisions": accessPublic,
	"ImmersionFeatureAccessGet":     accessAdmin,
	"ImmersionFeatureAccessGrant":   accessAdmin,
	"ImmersionFeatureAccessRevoke":  accessAdmin,

	"ProfileUsersList": accessAdmin,
}

var operationFixtures = map[string]string{
	"AuthzPermissionCheck":      "AuthzPermissionCheck",
	"AuthzProxyProxyAdminCheck": "AuthzProxyProxyAdminCheck",
	"AuthzRoleUpdate":           "AuthzRoleUpdate",

	"ContentAnnouncementCreate":   "CreateAnnouncement",
	"ContentAnnouncementDelete":   "DeleteAnnouncement",
	"ContentAnnouncementFindByID": "FindAnnouncementByID",
	"ContentAnnouncementList":     "ListAnnouncements",
	"ContentAnnouncementUpdate":   "UpdateAnnouncement",
	"ContentPageCreate":           "CreatePage",
	"ContentPageDelete":           "DeletePage",
	"ContentPageList":             "ListPages",
	"ContentPageUpdate":           "UpdatePage",
	"ContentPageVersionGet":       "GetPageVersion",
	"ContentPageVersionList":      "ListPageVersions",
	"ContentPostCreate":           "CreatePost",
	"ContentPostDelete":           "DeletePost",
	"ContentPostUpdate":           "UpdatePost",
	"ContentPostVersionGet":       "GetPostVersion",
	"ContentPostVersionList":      "ListPostVersions",

	"ImmersionContestCreate":                   "CreateContest",
	"ImmersionContestFindOngoingRegistrations": "FindOngoingContestRegistrations",
	"ImmersionContestFindRegistration":         "FindContestRegistration",
	"ImmersionContestRegistrationUpsert":       "UpsertContestRegistration",

	"ImmersionLogGetConfigurations":         "ImmersionLogGetConfigurations",
	"ImmersionLogTagSuggestions":            "ImmersionLogTagSuggestions",
	"ImmersionScorePreview":                 "ImmersionScorePreview",
	"ImmersionScoringRuleSetListPlatform":   "ImmersionScoringRuleSetListPlatform",
	"ImmersionScoringRuleSetListContest":    "ImmersionScoringRuleSetListContest",
	"ImmersionScoringRuleSetCreatePlatform": "ImmersionScoringRuleSetCreatePlatform",
	"ImmersionScoringRuleSetCreateContest":  "ImmersionScoringRuleSetCreateContest",

	"ImmersionLanguageList":        "ListLanguages",
	"ImmersionLanguageCreate":      "CreateLanguage",
	"ImmersionLanguageUpdate":      "UpdateLanguage",
	"ImmersionFeatureAccessGet":    "ImmersionFeatureAccessGet",
	"ImmersionFeatureAccessGrant":  "ImmersionFeatureAccessGrant",
	"ImmersionFeatureAccessRevoke": "ImmersionFeatureAccessRevoke",

	"ProfileUsersList": "ProfileUsersList",
}

func TestOperationAccessLevels(t *testing.T) {
	servers := []reflect.Type{
		reflect.TypeOf((*openapi.StrictServerInterface)(nil)).Elem(),
		reflect.TypeOf((*callbackopenapi.StrictServerInterface)(nil)).Elem(),
	}
	methods := make(map[string]bool)

	for _, server := range servers {
		for i := 0; i < server.NumMethod(); i++ {
			operation := server.Method(i).Name
			if methods[operation] {
				t.Errorf("operation %q is generated by multiple servers", operation)
				continue
			}
			methods[operation] = true

			level, ok := operationAccess[operation]
			if !ok {
				t.Errorf("operation %q has no declared access level", operation)
				continue
			}

			switch level {
			case accessPublic:
			case accessAuthenticated:
				requireAccessFixture(t, operation, http.StatusUnauthorized)
			case accessAdmin:
				requireAccessFixture(t, operation, http.StatusUnauthorized)
				requireAccessFixture(t, operation, http.StatusForbidden)
			case accessCallback:
				requireAccessFixture(t, operation, http.StatusUnauthorized)
			default:
				t.Errorf("operation %q has unknown access level %d", operation, level)
			}
		}
	}

	for operation := range operationAccess {
		if !methods[operation] {
			t.Errorf("declared operation %q is not on a generated strict server interface", operation)
		}
	}
}

func requireAccessFixture(t *testing.T, operation string, status int) {
	t.Helper()

	fixtureOperation, ok := operationFixtures[operation]
	if !ok {
		t.Errorf("protected operation %q has no fixture directory mapping", operation)
		return
	}

	directory := filepath.Join("testdata", fixtureOperation)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Errorf("read fixtures for operation %q: %v", operation, err)
		return
	}

	prefix := fmt.Sprintf("%d_", status)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			return
		}
	}

	t.Errorf("operation %q requires a %s* fixture directory under %s", operation, prefix, directory)
}
