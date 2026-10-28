package profile

import (
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestNormalizeUserPage(t *testing.T) {
	t.Parallel()

	t.Run("rejects negative page_size", func(t *testing.T) {
		t.Parallel()
		_, _, err := normalizeUserPage(-1, 0)
		if errx.KindOf(err) != errx.InvalidInput || err.Error() != "page_size must not be negative" {
			t.Errorf("error=%v, want invalid input %q", err, "page_size must not be negative")
		}
	})

	t.Run("rejects negative page", func(t *testing.T) {
		t.Parallel()
		_, _, err := normalizeUserPage(10, -1)
		if errx.KindOf(err) != errx.InvalidInput || err.Error() != "page must not be negative" {
			t.Errorf("error=%v, want invalid input %q", err, "page must not be negative")
		}
	})

	t.Run("defaults zero page_size to 20", func(t *testing.T) {
		t.Parallel()
		pageSize, page, err := normalizeUserPage(0, 0)
		if err != nil {
			t.Fatalf("error=%v, want nil", err)
		}
		if pageSize != 20 || page != 0 {
			t.Errorf("pageSize=%d page=%d, want 20 0", pageSize, page)
		}
	})

	t.Run("caps page_size at 100", func(t *testing.T) {
		t.Parallel()
		pageSize, page, err := normalizeUserPage(101, 2)
		if err != nil {
			t.Fatalf("error=%v, want nil", err)
		}
		if pageSize != 100 || page != 2 {
			t.Errorf("pageSize=%d page=%d, want 100 2", pageSize, page)
		}
	})
}
