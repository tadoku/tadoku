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
)

type accessLevel int

const (
	accessPublic accessLevel = iota
	accessAuthenticated
	accessAdmin
)

// Every operation on openapi.StrictServerInterface must declare its access level.
var operationAccess = map[string]accessLevel{
	"AuthzPermissionCheck":          accessAuthenticated,
	"AuthzRoleGet":                  accessPublic,
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
}

var operationFixtures = map[string]string{
	"AuthzPermissionCheck":        "AuthzPermissionCheck",
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
}

func TestOperationAccessLevels(t *testing.T) {
	server := reflect.TypeOf((*openapi.StrictServerInterface)(nil)).Elem()
	methods := make(map[string]bool, server.NumMethod())

	for i := 0; i < server.NumMethod(); i++ {
		operation := server.Method(i).Name
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
		default:
			t.Errorf("operation %q has unknown access level %d", operation, level)
		}
	}

	for operation := range operationAccess {
		if !methods[operation] {
			t.Errorf("declared operation %q is not a method on openapi.StrictServerInterface", operation)
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
