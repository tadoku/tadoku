package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMigration struct {
	upErr            error
	sourceCloseErr   error
	databaseCloseErr error
	upCalled         bool
	closeCalled      bool
}

func setPostgresEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_HOST", "db")
	t.Setenv("POSTGRES_DATABASE", "database")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "password")
	t.Setenv("POSTGRES_SSLMODE", "require")
}

func (m *fakeMigration) Up() error {
	m.upCalled = true
	return m.upErr
}

func (m *fakeMigration) Close() (error, error) {
	m.closeCalled = true
	return m.sourceCloseErr, m.databaseCloseErr
}

func TestRunSuccessfulMigration(t *testing.T) {
	setPostgresEnvironment(t)
	runner := &fakeMigration{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{"-source", "file:///migrations", "up"},
		&stdout,
		&stderr,
		func(sourceURL, databaseURL string) (migration, error) {
			assert.Equal(t, "file:///migrations", sourceURL)
			assert.Equal(t, "postgres://user:password@db:5432/database?sslmode=require", databaseURL)
			return runner, nil
		},
	)

	assert.Equal(t, 0, exitCode)
	assert.True(t, runner.upCalled)
	assert.True(t, runner.closeCalled)
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestRunNoChangeSucceeds(t *testing.T) {
	setPostgresEnvironment(t)
	runner := &fakeMigration{upErr: migrate.ErrNoChange}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{"-source", "file:///migrations", "up"},
		&stdout,
		&stderr,
		func(string, string) (migration, error) {
			return runner, nil
		},
	)

	assert.Equal(t, 0, exitCode)
	assert.True(t, runner.closeCalled)
	assert.Equal(t, "no change\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestRunMigrationFailureFailsAndCloses(t *testing.T) {
	setPostgresEnvironment(t)
	runner := &fakeMigration{upErr: errors.New("migration failed")}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{"-source", "file:///migrations", "up"},
		&stdout,
		&stderr,
		func(string, string) (migration, error) {
			return runner, nil
		},
	)

	assert.Equal(t, 1, exitCode)
	assert.True(t, runner.closeCalled)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "migration failed")
}

func TestRunRejectsInvalidArgumentsBeforeInitialization(t *testing.T) {
	var stderr bytes.Buffer
	factoryCalled := false

	exitCode := run(
		[]string{"-source", "file:///migrations", "down"},
		&bytes.Buffer{},
		&stderr,
		func(string, string) (migration, error) {
			factoryCalled = true
			return nil, nil
		},
	)

	assert.Equal(t, 2, exitCode)
	assert.False(t, factoryCalled)
	assert.Contains(t, stderr.String(), "expected command: up")
}

func TestRunRejectsRemovedDatabaseFlag(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := run(
		[]string{"-source", "file:///migrations", "-database", "postgres://database", "up"},
		&bytes.Buffer{},
		&stderr,
		func(string, string) (migration, error) { return nil, nil },
	)

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stderr.String(), "flag provided but not defined: -database")
}

func TestRunUsesIndividualPostgresEnvironment(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "db")
	t.Setenv("POSTGRES_DATABASE", "database")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "sentinel:/?#")
	t.Setenv("POSTGRES_SSLMODE", "require")
	var stderr bytes.Buffer

	exitCode := run(
		[]string{"-source", "file:///migrations", "up"},
		&bytes.Buffer{},
		&stderr,
		func(_, databaseURL string) (migration, error) {
			assert.Contains(t, databaseURL, "sentinel%3A%2F%3F%23")
			return nil, errors.New("sentinel:/?#")
		},
	)

	assert.Equal(t, 1, exitCode)
	assert.NotContains(t, stderr.String(), "sentinel")
}

func TestRunInitializationFailure(t *testing.T) {
	setPostgresEnvironment(t)
	var stderr bytes.Buffer
	expectedErr := errors.New("cannot initialize")

	exitCode := run(
		[]string{"-source", "file:///migrations", "up"},
		&bytes.Buffer{},
		&stderr,
		func(string, string) (migration, error) {
			return nil, expectedErr
		},
	)

	require.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), expectedErr.Error())
}
