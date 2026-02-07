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
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockUpdateUserRoleRepo struct {
	userExists bool
	existsErr  error

	auditErr  error
	auditReq  *domain.ModerationAuditLogCreateRequest
	auditCall bool
}

func (m *mockUpdateUserRoleRepo) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.userExists, m.existsErr
}

func (m *mockUpdateUserRoleRepo) CreateModerationAuditLog(ctx context.Context, req *domain.ModerationAuditLogCreateRequest) error {
	m.auditCall = true
	m.auditReq = req
	return m.auditErr
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

func TestUpdateUserRole_Execute(t *testing.T) {
	targetID := uuid.New()

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		repo := &mockUpdateUserRoleRepo{}
		svc := domain.NewUpdateUserRole(repo, repo, &mockClaimsService{}, &mockRoleManager{})

		err := svc.Execute(ctxWithRole(commondomain.RoleGuest), &domain.UpdateUserRoleRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("returns forbidden for non-admin", func(t *testing.T) {
		repo := &mockUpdateUserRoleRepo{}
		svc := domain.NewUpdateUserRole(repo, repo, &mockClaimsService{}, &mockRoleManager{})

		err := svc.Execute(ctxWithRole(commondomain.RoleUser), &domain.UpdateUserRoleRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("bans a user via keto and audits", func(t *testing.T) {
		repo := &mockUpdateUserRoleRepo{userExists: true}
		roleMgmt := &mockRoleManager{}
		rolesSvc := &mockClaimsService{
			claims: map[string]commonroles.Claims{
				targetID.String(): {Subject: targetID.String(), Authenticated: true, Admin: false, Banned: false},
			},
		}
		svc := domain.NewUpdateUserRole(repo, repo, rolesSvc, roleMgmt)

		ctx := ctxWithRole(commondomain.RoleAdmin)
		err := svc.Execute(ctx, &domain.UpdateUserRoleRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		require.NoError(t, err)
		assert.True(t, roleMgmt.setBannedCalled)
		assert.Equal(t, targetID.String(), roleMgmt.setBannedSubj)
		assert.True(t, roleMgmt.setBannedVal)
		assert.True(t, repo.auditCall)
		require.NotNil(t, repo.auditReq)
		assert.Equal(t, uuid.MustParse(testSubjectID), repo.auditReq.ModeratorUserID)
	})

	t.Run("unbans a user via keto and audits", func(t *testing.T) {
		repo := &mockUpdateUserRoleRepo{userExists: true}
		roleMgmt := &mockRoleManager{}
		rolesSvc := &mockClaimsService{
			claims: map[string]commonroles.Claims{
				targetID.String(): {Subject: targetID.String(), Authenticated: true, Admin: false, Banned: true},
			},
		}
		svc := domain.NewUpdateUserRole(repo, repo, rolesSvc, roleMgmt)

		ctx := ctxWithRole(commondomain.RoleAdmin)
		err := svc.Execute(ctx, &domain.UpdateUserRoleRequest{
			UserID: targetID,
			Role:   "user",
			Reason: "reason",
		})

		require.NoError(t, err)
		assert.True(t, roleMgmt.setBannedCalled)
		assert.Equal(t, targetID.String(), roleMgmt.setBannedSubj)
		assert.False(t, roleMgmt.setBannedVal)
		assert.True(t, repo.auditCall)
	})

	t.Run("cannot modify role of an admin target", func(t *testing.T) {
		repo := &mockUpdateUserRoleRepo{userExists: true}
		roleMgmt := &mockRoleManager{}
		rolesSvc := &mockClaimsService{
			claims: map[string]commonroles.Claims{
				targetID.String(): {Subject: targetID.String(), Authenticated: true, Admin: true},
			},
		}
		svc := domain.NewUpdateUserRole(repo, repo, rolesSvc, roleMgmt)

		err := svc.Execute(ctxWithRole(commondomain.RoleAdmin), &domain.UpdateUserRoleRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.False(t, roleMgmt.setBannedCalled)
	})

	t.Run("returns authz unavailable when keto read fails", func(t *testing.T) {
		repo := &mockUpdateUserRoleRepo{userExists: true}
		roleMgmt := &mockRoleManager{}
		rolesSvc := &mockClaimsService{err: errors.New("keto down")}
		svc := domain.NewUpdateUserRole(repo, repo, rolesSvc, roleMgmt)

		err := svc.Execute(ctxWithRole(commondomain.RoleAdmin), &domain.UpdateUserRoleRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, domain.ErrAuthzUnavailable)
	})

	t.Run("returns authz unavailable when keto write fails", func(t *testing.T) {
		repo := &mockUpdateUserRoleRepo{userExists: true}
		roleMgmt := &mockRoleManager{setBannedErr: errors.New("keto write down")}
		rolesSvc := &mockClaimsService{
			claims: map[string]commonroles.Claims{
				targetID.String(): {Subject: targetID.String(), Authenticated: true, Admin: false},
			},
		}
		svc := domain.NewUpdateUserRole(repo, repo, rolesSvc, roleMgmt)

		err := svc.Execute(ctxWithRole(commondomain.RoleAdmin), &domain.UpdateUserRoleRequest{
			UserID: targetID,
			Role:   "banned",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, domain.ErrAuthzUnavailable)
	})

	t.Run("validates role", func(t *testing.T) {
		repo := &mockUpdateUserRoleRepo{}
		svc := domain.NewUpdateUserRole(repo, repo, &mockClaimsService{}, &mockRoleManager{})

		err := svc.Execute(ctxWithRole(commondomain.RoleAdmin), &domain.UpdateUserRoleRequest{
			UserID: targetID,
			Role:   "nope",
			Reason: "reason",
		})

		assert.ErrorIs(t, err, domain.ErrRequestInvalid)
	})
}
