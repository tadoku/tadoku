package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	commondomain "github.com/tadoku/tadoku/services/common/domain"

	"github.com/tadoku/tadoku/services/authz-api/domain"
)

type mockUserDir struct {
	exists bool
	err    error
}

func (m *mockUserDir) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.exists, m.err
}

type mockAuditRepo struct {
	called bool
	req    *domain.ModerationAuditLogCreateRequest
	err    error
}

func (m *mockAuditRepo) CreateModerationAuditLog(ctx context.Context, req *domain.ModerationAuditLogCreateRequest) error {
	m.called = true
	m.req = req
	return m.err
}

type mockRoleManager struct {
	setBannedCalled bool
	setBannedSubj   string
	setBannedVal    bool
	setBannedErr    error
}

func (m *mockRoleManager) SetAdmin(ctx context.Context, subjectID string, enabled bool) error {
	return errors.New("not implemented")
}

func (m *mockRoleManager) SetBanned(ctx context.Context, subjectID string, enabled bool) error {
	m.setBannedCalled = true
	m.setBannedSubj = subjectID
	m.setBannedVal = enabled
	return m.setBannedErr
}

type mockClaimsService struct {
	claims map[string]commonroles.Claims
	err    error
}

func (m *mockClaimsService) ClaimsForSubject(ctx context.Context, subjectID string) (commonroles.Claims, error) {
	if m.err != nil {
		return commonroles.Claims{Subject: subjectID, Authenticated: true, Err: m.err}, m.err
	}
	if c, ok := m.claims[subjectID]; ok {
		return c, nil
	}
	return commonroles.Claims{Subject: subjectID, Authenticated: true}, nil
}

func (m *mockClaimsService) ClaimsForSubjects(ctx context.Context, subjectIDs []string) (map[string]commonroles.Claims, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make(map[string]commonroles.Claims, len(subjectIDs))
	for _, id := range subjectIDs {
		c, err := m.ClaimsForSubject(ctx, id)
		if err != nil {
			return nil, err
		}
		out[id] = c
	}
	return out, nil
}

func ctxWithUserSubject(subject string) context.Context {
	return context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.UserIdentity{Subject: subject})
}

func TestRoleUpdate_Execute(t *testing.T) {
	moderatorID := uuid.New()
	targetID := uuid.New()

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		users := &mockUserDir{}
		audit := &mockAuditRepo{}
		svc := domain.NewRoleUpdate(users, audit, &mockClaimsService{}, &mockRoleManager{})

		err := svc.Execute(ctxWithUserSubject("guest"), &domain.RoleUpdateRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, commondomain.ErrUnauthorized)
	})

	t.Run("returns forbidden for non-admin moderator", func(t *testing.T) {
		users := &mockUserDir{}
		audit := &mockAuditRepo{}
		rolesSvc := &mockClaimsService{
			claims: map[string]commonroles.Claims{
				moderatorID.String(): {Subject: moderatorID.String(), Authenticated: true, Admin: false},
			},
		}
		svc := domain.NewRoleUpdate(users, audit, rolesSvc, &mockRoleManager{})

		err := svc.Execute(ctxWithUserSubject(moderatorID.String()), &domain.RoleUpdateRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, commondomain.ErrForbidden)
	})

	t.Run("bans a user via keto and audits", func(t *testing.T) {
		users := &mockUserDir{exists: true}
		audit := &mockAuditRepo{}
		roleMgmt := &mockRoleManager{}
		rolesSvc := &mockClaimsService{
			claims: map[string]commonroles.Claims{
				moderatorID.String(): {Subject: moderatorID.String(), Authenticated: true, Admin: true},
				targetID.String():    {Subject: targetID.String(), Authenticated: true, Admin: false, Banned: false},
			},
		}
		svc := domain.NewRoleUpdate(users, audit, rolesSvc, roleMgmt)

		err := svc.Execute(ctxWithUserSubject(moderatorID.String()), &domain.RoleUpdateRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		require.NoError(t, err)
		assert.True(t, roleMgmt.setBannedCalled)
		assert.Equal(t, targetID.String(), roleMgmt.setBannedSubj)
		assert.True(t, roleMgmt.setBannedVal)
		assert.True(t, audit.called)
		require.NotNil(t, audit.req)
		assert.Equal(t, moderatorID, audit.req.ModeratorUserID)
		assert.Equal(t, "ban_user", audit.req.Action)
	})

	t.Run("unbans a user via keto and audits", func(t *testing.T) {
		users := &mockUserDir{exists: true}
		audit := &mockAuditRepo{}
		roleMgmt := &mockRoleManager{}
		rolesSvc := &mockClaimsService{
			claims: map[string]commonroles.Claims{
				moderatorID.String(): {Subject: moderatorID.String(), Authenticated: true, Admin: true},
				targetID.String():    {Subject: targetID.String(), Authenticated: true, Admin: false, Banned: true},
			},
		}
		svc := domain.NewRoleUpdate(users, audit, rolesSvc, roleMgmt)

		err := svc.Execute(ctxWithUserSubject(moderatorID.String()), &domain.RoleUpdateRequest{
			UserID: targetID,
			Role:   "user",
			Reason: "reason",
		})

		require.NoError(t, err)
		assert.True(t, roleMgmt.setBannedCalled)
		assert.False(t, roleMgmt.setBannedVal)
		assert.True(t, audit.called)
		require.NotNil(t, audit.req)
		assert.Equal(t, "unban_user", audit.req.Action)
	})

	t.Run("cannot modify role of an admin target", func(t *testing.T) {
		users := &mockUserDir{exists: true}
		audit := &mockAuditRepo{}
		roleMgmt := &mockRoleManager{}
		rolesSvc := &mockClaimsService{
			claims: map[string]commonroles.Claims{
				moderatorID.String(): {Subject: moderatorID.String(), Authenticated: true, Admin: true},
				targetID.String():    {Subject: targetID.String(), Authenticated: true, Admin: true},
			},
		}
		svc := domain.NewRoleUpdate(users, audit, rolesSvc, roleMgmt)

		err := svc.Execute(ctxWithUserSubject(moderatorID.String()), &domain.RoleUpdateRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, commondomain.ErrForbidden)
		assert.False(t, roleMgmt.setBannedCalled)
	})

	t.Run("returns not found when user does not exist", func(t *testing.T) {
		users := &mockUserDir{exists: false}
		audit := &mockAuditRepo{}
		roleMgmt := &mockRoleManager{}
		rolesSvc := &mockClaimsService{
			claims: map[string]commonroles.Claims{
				moderatorID.String(): {Subject: moderatorID.String(), Authenticated: true, Admin: true},
			},
		}
		svc := domain.NewRoleUpdate(users, audit, rolesSvc, roleMgmt)

		err := svc.Execute(ctxWithUserSubject(moderatorID.String()), &domain.RoleUpdateRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, commondomain.ErrNotFound)
	})
}

