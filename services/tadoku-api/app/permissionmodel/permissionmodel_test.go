package permissionmodel

import "testing"

func TestNewBuildsTheProductionModel(t *testing.T) {
	t.Parallel()

	if _, err := New(); err != nil {
		t.Error(err)
	}
}
