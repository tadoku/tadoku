package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/testutil/authzctx"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/postgrestest"
)

func TestLogCreateCommitsAndReturnsLogAgainstMigratedSchema(t *testing.T) {
	testDB := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(testDB)
	now := time.Date(2026, 8, 30, 7, 0, 0, 0, time.UTC)
	userID := uuid.New()
	ctx := authzctx.UserIdentity(userID.String(), "Submission Regression User")
	durationSeconds := int32(3600)

	// TODO: Cover this through the public POST /logs endpoint in the end-to-end
	// suite so authentication, request mapping, commit, and response rendering
	// are exercised together.
	service := domain.NewLogCreate(repo, commondomain.NewMockClock(now), domain.NewUserUpsert(repo))
	created, err := service.Execute(ctx, &domain.LogCreateRequest{
		ActivityID:      1,
		LanguageCode:    "eng",
		DurationSeconds: &durationSeconds,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, userID, created.UserID)
	assert.Equal(t, "Submission Regression User", *created.UserDisplayName)
	assert.Equal(t, float32(12), created.Score)

	var persistedLogs int
	require.NoError(t, testDB.QueryRow(
		"select count(*) from logs where id = $1 and user_id = $2",
		created.ID,
		userID,
	).Scan(&persistedLogs))
	assert.Equal(t, 1, persistedLogs)
}
