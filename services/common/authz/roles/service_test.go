package roles

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
)

type fakeKeto struct {
	results         map[string]ketoclient.PermissionResult // keyed by relation
	subjectIDsByRel map[string][]string                    // keyed by relation
	listSubjectsErr error
}

func (f *fakeKeto) CheckPermission(ctx context.Context, namespace, object, relation string, subject ketoclient.Subject) (bool, error) {
	r, ok := f.results[relation]
	if !ok {
		return false, errors.New("missing relation in fake")
	}
	return r.Allowed, r.Err
}

func (f *fakeKeto) CheckPermissions(ctx context.Context, checks []ketoclient.PermissionCheck) []ketoclient.PermissionResult {
	out := make([]ketoclient.PermissionResult, 0, len(checks))
	for _, c := range checks {
		r, ok := f.results[c.Relation]
		if !ok {
			out = append(out, ketoclient.PermissionResult{Check: c, Allowed: false, Err: errors.New("missing relation in fake")})
			continue
		}
		out = append(out, ketoclient.PermissionResult{Check: c, Allowed: r.Allowed, Err: r.Err})
	}
	return out
}

func (f *fakeKeto) ListSubjectIDsForRelation(ctx context.Context, namespace, object, relation string) ([]string, error) {
	if f.listSubjectsErr != nil {
		return nil, f.listSubjectsErr
	}
	return f.subjectIDsByRel[relation], nil
}

func TestKetoService_ClaimsForSubject_Guest(t *testing.T) {
	svc := NewKetoService(&fakeKeto{}, "app", "tadoku")
	claims, err := svc.ClaimsForSubject(context.Background(), "guest")
	require.NoError(t, err)
	assert.False(t, claims.Authenticated)
	assert.False(t, claims.Admin)
	assert.False(t, claims.Banned)
}

func TestKetoService_ClaimsForSubject_Admin(t *testing.T) {
	svc := NewKetoService(&fakeKeto{
		results: map[string]ketoclient.PermissionResult{
			"admins": {Allowed: true},
			"banned": {Allowed: false},
		},
	}, "app", "tadoku")

	claims, err := svc.ClaimsForSubject(context.Background(), "kratos-id")
	require.NoError(t, err)
	assert.True(t, claims.Authenticated)
	assert.True(t, claims.Admin)
	assert.False(t, claims.Banned)
}

func TestKetoService_ClaimsForSubject_Banned(t *testing.T) {
	svc := NewKetoService(&fakeKeto{
		results: map[string]ketoclient.PermissionResult{
			"admins": {Allowed: false},
			"banned": {Allowed: true},
		},
	}, "app", "tadoku")

	claims, err := svc.ClaimsForSubject(context.Background(), "kratos-id")
	require.NoError(t, err)
	assert.True(t, claims.Authenticated)
	assert.False(t, claims.Admin)
	assert.True(t, claims.Banned)
}

func TestKetoService_ClaimsForSubject_Error(t *testing.T) {
	svc := NewKetoService(&fakeKeto{
		results: map[string]ketoclient.PermissionResult{
			"admins": {Allowed: false, Err: errors.New("boom")},
			"banned": {Allowed: false},
		},
	}, "app", "tadoku")

	claims, err := svc.ClaimsForSubject(context.Background(), "kratos-id")
	require.Error(t, err)
	assert.Error(t, claims.Err)
}

func TestKetoService_ClaimsForSubjects(t *testing.T) {
	svc := NewKetoService(&fakeKeto{
		subjectIDsByRel: map[string][]string{
			"admins": {"a"},
			"banned": {"b"},
		},
	}, "app", "tadoku")

	claimsBySubject, err := svc.ClaimsForSubjects(context.Background(), []string{"a", "b", "c", "guest", ""})
	require.NoError(t, err)

	assert.True(t, claimsBySubject["a"].Authenticated)
	assert.True(t, claimsBySubject["a"].Admin)
	assert.False(t, claimsBySubject["a"].Banned)

	assert.True(t, claimsBySubject["b"].Authenticated)
	assert.False(t, claimsBySubject["b"].Admin)
	assert.True(t, claimsBySubject["b"].Banned)

	assert.True(t, claimsBySubject["c"].Authenticated)
	assert.False(t, claimsBySubject["c"].Admin)
	assert.False(t, claimsBySubject["c"].Banned)

	assert.False(t, claimsBySubject["guest"].Authenticated)
	assert.False(t, claimsBySubject[""].Authenticated)
}
