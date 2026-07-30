package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockScoringRuleSetManagementRepository struct {
	contest          *domain.ContestView
	ruleSet          *domain.ScoringRuleSet
	createdWith      *domain.ScoringRuleSetDraftCreateRequest
	published        bool
	activatedContest uuid.UUID
	activatedRuleSet uuid.UUID
}

func (m *mockScoringRuleSetManagementRepository) FindContestByID(context.Context, *domain.ContestFindRequest) (*domain.ContestView, error) {
	return m.contest, nil
}

func (m *mockScoringRuleSetManagementRepository) FindScoringRuleSetByID(context.Context, uuid.UUID) (*domain.ScoringRuleSet, error) {
	return m.ruleSet, nil
}

func (m *mockScoringRuleSetManagementRepository) ListPlatformScoringRuleSets(context.Context) ([]domain.ScoringRuleSet, error) {
	return nil, nil
}

func (m *mockScoringRuleSetManagementRepository) ListLanguages(context.Context) ([]domain.Language, error) {
	return []domain.Language{{Code: "jpn", Name: "Japanese"}}, nil
}

func (m *mockScoringRuleSetManagementRepository) ListContestScoringRuleSets(context.Context, uuid.UUID) ([]domain.ScoringRuleSet, error) {
	return nil, nil
}

func (m *mockScoringRuleSetManagementRepository) CreateScoringRuleSetDraft(_ context.Context, req *domain.ScoringRuleSetDraftCreateRequest) (*domain.ScoringRuleSet, error) {
	m.createdWith = req
	return &domain.ScoringRuleSet{ID: uuid.New(), Status: domain.ScoringRuleSetStatusDraft}, nil
}

func (m *mockScoringRuleSetManagementRepository) PublishScoringRuleSet(context.Context, uuid.UUID, time.Time) (*domain.ScoringRuleSet, error) {
	m.published = true
	return &domain.ScoringRuleSet{Status: domain.ScoringRuleSetStatusPublished}, nil
}

func (m *mockScoringRuleSetManagementRepository) ActivatePlatformScoringRuleSet(_ context.Context, ruleSetID uuid.UUID) error {
	m.activatedRuleSet = ruleSetID
	return nil
}

func (m *mockScoringRuleSetManagementRepository) ActivateContestScoringRuleSet(
	_ context.Context,
	contestID uuid.UUID,
	ruleSetID uuid.UUID,
	_ time.Time,
) error {
	m.activatedContest = contestID
	m.activatedRuleSet = ruleSetID
	return nil
}

func TestScoringRuleSetManagementCreatesNormalizedContestDraft(t *testing.T) {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	contestID := uuid.New()
	fallbackID := uuid.New()
	repo := &mockScoringRuleSetManagementRepository{
		contest: &domain.ContestView{
			ID:           contestID,
			OwnerUserID:  userID,
			ContestStart: now.Add(time.Hour),
		},
		ruleSet: &domain.ScoringRuleSet{
			ID:     fallbackID,
			Scope:  domain.ScoringRuleSetScopePlatform,
			Status: domain.ScoringRuleSetStatusPublished,
		},
	}
	service := domain.NewScoringRuleSetManagement(repo, commondomain.NewMockClock(now))

	_, err := service.CreateContestDraft(
		ctxWithUserSubject(userID.String()),
		contestID,
		&domain.ScoringRuleSetDraftCreateRequest{
			Mode:              domain.ScoringRuleSetModeOverride,
			FallbackRuleSetID: &fallbackID,
			Rules: []domain.ScoringRule{{
				Priority:     1,
				ActivityID:   1,
				UnitKey:      domain.UnitKeyReadingPage,
				LanguageCode: " JPN ",
				Tag:          " Book ",
				ScoreSource:  domain.ScoreSourceAmount,
				Rate:         2,
			}},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, repo.createdWith)
	assert.Equal(t, domain.ScoringRuleSetScopeContest, repo.createdWith.Scope())
	assert.Equal(t, &contestID, repo.createdWith.ContestID)
	assert.Equal(t, "jpn", repo.createdWith.Rules[0].LanguageCode)
	assert.Equal(t, "book", repo.createdWith.Rules[0].Tag)
}

func TestScoringRuleSetManagementRejectsMismatchedUnit(t *testing.T) {
	repo := &mockScoringRuleSetManagementRepository{}
	service := domain.NewScoringRuleSetManagement(repo, commondomain.NewMockClock(time.Now()))

	_, err := service.CreatePlatformDraft(ctxWithAdmin(), &domain.ScoringRuleSetDraftCreateRequest{
		Rules: []domain.ScoringRule{{
			Priority:    1,
			ActivityID:  2,
			UnitKey:     domain.UnitKeyReadingPage,
			ScoreSource: domain.ScoreSourceAmount,
			Rate:        1,
		}},
	})

	assert.ErrorIs(t, err, domain.ErrInvalidScoringRuleSet)
	assert.Nil(t, repo.createdWith)
}

func TestScoringRuleSetManagementRejectsContestChangeAfterStart(t *testing.T) {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	contestID := uuid.New()
	repo := &mockScoringRuleSetManagementRepository{
		contest: &domain.ContestView{
			ID:           contestID,
			OwnerUserID:  userID,
			ContestStart: now,
		},
	}
	service := domain.NewScoringRuleSetManagement(repo, commondomain.NewMockClock(now))

	_, err := service.CreateContestDraft(
		ctxWithUserSubject(userID.String()),
		contestID,
		&domain.ScoringRuleSetDraftCreateRequest{Mode: domain.ScoringRuleSetModeReplace},
	)

	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestScoringRuleSetManagementActivatesPublishedContestVersion(t *testing.T) {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	contestID := uuid.New()
	ruleSetID := uuid.New()
	repo := &mockScoringRuleSetManagementRepository{
		contest: &domain.ContestView{
			ID:           contestID,
			OwnerUserID:  userID,
			ContestStart: now.Add(time.Hour),
		},
		ruleSet: &domain.ScoringRuleSet{
			ID:        ruleSetID,
			Scope:     domain.ScoringRuleSetScopeContest,
			ContestID: &contestID,
			Status:    domain.ScoringRuleSetStatusPublished,
		},
	}
	service := domain.NewScoringRuleSetManagement(repo, commondomain.NewMockClock(now))

	err := service.Activate(ctxWithUserSubject(userID.String()), ruleSetID)

	require.NoError(t, err)
	assert.Equal(t, contestID, repo.activatedContest)
	assert.Equal(t, ruleSetID, repo.activatedRuleSet)
}

func TestScoringRuleSetManagementDoesNotRepublishPublishedVersion(t *testing.T) {
	ruleSetID := uuid.New()
	repo := &mockScoringRuleSetManagementRepository{
		ruleSet: &domain.ScoringRuleSet{
			ID:     ruleSetID,
			Scope:  domain.ScoringRuleSetScopePlatform,
			Status: domain.ScoringRuleSetStatusPublished,
		},
	}
	service := domain.NewScoringRuleSetManagement(repo, commondomain.NewMockClock(time.Now()))

	_, err := service.Publish(ctxWithAdmin(), ruleSetID)

	assert.ErrorIs(t, err, domain.ErrConflict)
	assert.False(t, repo.published)
}
