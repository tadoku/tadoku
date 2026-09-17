package spec_test

import (
	"context"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
)

func TestMergedContractIsValidOpenAPI(t *testing.T) {
	contract := loadContract(t)
	if err := contract.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestAnnouncementListBoundMatchesServerClamp(t *testing.T) {
	contract := loadContract(t)
	announcementList := contract.Components.Schemas["ContentAnnouncementList"].Value
	announcements := announcementList.AllOf[1].Value.Properties["announcements"].Value
	if announcements.MaxItems == nil {
		t.Fatal("ContentAnnouncementList announcements has no maxItems")
	}

	if got, want := *announcements.MaxItems, uint64(content.AnnouncementListMaxPageSize); got != want {
		t.Errorf("ContentAnnouncementList announcements maxItems = %d, server clamp = %d", got, want)
	}
}

func loadContract(t *testing.T) *openapi3.T {
	t.Helper()

	path, err := bazel.Runfile("services/tadoku-api/spec/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	contract, err := openapi3.NewLoader().LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return contract
}
