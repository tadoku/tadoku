package httperr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
		ok   bool
	}{
		{"nil", nil, 0, false},
		{"wrapped request invalid", fmt.Errorf("wrap: %w", commondomain.ErrRequestInvalid), http.StatusBadRequest, true},
		{"unauthorized", commondomain.ErrUnauthorized, http.StatusUnauthorized, true},
		{"forbidden", commondomain.ErrForbidden, http.StatusForbidden, true},
		{"not found", commondomain.ErrNotFound, http.StatusNotFound, true},
		{"conflict", commondomain.ErrConflict, http.StatusConflict, true},
		{"authz unavailable", commondomain.ErrAuthzUnavailable, http.StatusServiceUnavailable, true},
		{"unknown", errors.New("nope"), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := StatusCode(tt.err)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}
