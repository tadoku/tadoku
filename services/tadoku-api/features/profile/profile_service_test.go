package profile

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
)

type recordingUserCache struct {
	users []CachedUser
	calls int
}

func (c *recordingUserCache) Users(context.Context) ([]CachedUser, error) {
	c.calls++
	return append([]CachedUser(nil), c.users...), nil
}

type recordingRoleProvider struct {
	claims map[string]commonroles.Claims
	err    error
	calls  [][]string
}

func (p *recordingRoleProvider) ClaimsForSubjects(_ context.Context, subjectIDs []string) (map[string]commonroles.Claims, error) {
	p.calls = append(p.calls, append([]string(nil), subjectIDs...))
	if p.err != nil {
		return nil, p.err
	}
	return p.claims, nil
}

func TestListUsersPaginationPreservesProviderOrder(t *testing.T) {
	users := make([]CachedUser, 105)
	for i := range users {
		users[i] = CachedUser{ID: fmt.Sprintf("user-%03d", i), DisplayName: fmt.Sprintf("User %03d", i)}
	}

	for _, test := range []struct {
		name      string
		pageSize  int
		page      int
		wantCount int
		wantFirst string
		wantLast  string
	}{
		{name: "default page size", pageSize: 0, wantCount: 20, wantFirst: "user-000", wantLast: "user-019"},
		{name: "negative page starts at zero", pageSize: 2, page: -4, wantCount: 2, wantFirst: "user-000", wantLast: "user-001"},
		{name: "page size is capped", pageSize: 1000, wantCount: 100, wantFirst: "user-000", wantLast: "user-099"},
		{name: "later page", pageSize: 20, page: 5, wantCount: 5, wantFirst: "user-100", wantLast: "user-104"},
		{name: "empty page", pageSize: 20, page: 6, wantCount: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			cache := &recordingUserCache{users: users}
			roles := &recordingRoleProvider{}
			result, err := NewService(cache, roles).ListUsers(t.Context(), test.pageSize, test.page, "")
			if err != nil {
				t.Fatal(err)
			}
			if result.TotalSize != len(users) || len(result.Users) != test.wantCount {
				t.Fatalf("total/page size = %d/%d, want %d/%d", result.TotalSize, len(result.Users), len(users), test.wantCount)
			}
			if test.wantCount > 0 {
				if result.Users[0].ID != test.wantFirst || result.Users[len(result.Users)-1].ID != test.wantLast {
					t.Errorf("page bounds = %q..%q, want %q..%q", result.Users[0].ID, result.Users[len(result.Users)-1].ID, test.wantFirst, test.wantLast)
				}
			}
		})
	}
}

func TestListUsersSearchesBeforePaginationAndReadsCurrentRoles(t *testing.T) {
	cache := &recordingUserCache{users: []CachedUser{
		{ID: "bobby", DisplayName: "Bobby", Email: "first@example.test"},
		{ID: "bob", DisplayName: "Bob", Email: "second@example.test"},
		{ID: "bob-user", DisplayName: "Bob User", Email: "user@example.test"},
		{ID: "alice", DisplayName: "Alice", Email: "alice@example.test"},
	}}
	roles := &recordingRoleProvider{claims: map[string]commonroles.Claims{
		"bob":   {Banned: true},
		"bobby": {Admin: true, Banned: true},
	}}
	ctx := t.Context()
	service := NewService(cache, roles)

	result, err := service.ListUsers(ctx, 2, 0, "BOB")
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalSize != 3 || len(result.Users) != 2 {
		t.Fatalf("search total/page size = %d/%d, want 3/2", result.TotalSize, len(result.Users))
	}
	if result.Users[0].ID != "bob" || result.Users[0].Role != "banned" {
		t.Errorf("highest relevance user = %+v, want bob with banned role", result.Users[0])
	}
	if len(roles.calls) != 1 || !reflect.DeepEqual(roles.calls[0], []string{"bob", "bobby"}) {
		t.Fatalf("role subjects = %v, want searched page only", roles.calls)
	}
	if result.Users[1].Role != "admin" {
		t.Errorf("admin+banned role = %q, want admin precedence", result.Users[1].Role)
	}

	roles.claims["bob"] = commonroles.Claims{}
	result, err = service.ListUsers(ctx, 1, 0, "BOB")
	if err != nil {
		t.Fatal(err)
	}
	if result.Users[0].Role != "user" {
		t.Errorf("second request role = %q, want freshly read user role", result.Users[0].Role)
	}
}
