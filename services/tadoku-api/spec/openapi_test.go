package spec_test

import (
	"context"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestMergedContractIsValidOpenAPI(t *testing.T) {
	path, err := bazel.Runfile("services/tadoku-api/spec/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	contract, err := openapi3.NewLoader().LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := contract.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
}
