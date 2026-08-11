package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/assert"
)

type fakeMigration struct {
	version            uint
	dirty              bool
	versionErr         error
	forceErr           error
	sourceCloseErr     error
	databaseCloseErr   error
	forcedVersion      int
	forceCalled        bool
	closeCalled        bool
	forceDoesNotUpdate bool
}

func (m *fakeMigration) Version() (uint, bool, error) {
	return m.version, m.dirty, m.versionErr
}

func (m *fakeMigration) Force(version int) error {
	m.forceCalled = true
	m.forcedVersion = version
	if m.forceErr == nil && !m.forceDoesNotUpdate {
		m.dirty = false
		if version == -1 {
			m.version = 0
			m.versionErr = migrate.ErrNilVersion
		} else {
			m.version = uint(version)
			m.versionErr = nil
		}
	}
	return m.forceErr
}

func (m *fakeMigration) Close() (error, error) {
	m.closeCalled = true
	return m.sourceCloseErr, m.databaseCloseErr
}

func runWithFake(t *testing.T, args []string, runner *fakeMigration) (int, string, string) {
	t.Helper()
	t.Setenv("POSTGRES_HOST", "db")
	t.Setenv("POSTGRES_DATABASE", "database")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "password")
	t.Setenv("POSTGRES_SSLMODE", "require")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		args,
		&stdout,
		&stderr,
		func(sourceURL, databaseURL string) (migration, error) {
			assert.Equal(t, "file:///migrations", sourceURL)
			assert.Equal(t, "postgres://user:password@db:5432/database?sslmode=require", databaseURL)
			return runner, nil
		},
	)
	return exitCode, stdout.String(), stderr.String()
}

func TestInspectCleanVersion(t *testing.T) {
	runner := &fakeMigration{version: 12}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{"-source", "file:///migrations", "inspect"},
		runner,
	)

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "version=12 dirty=false\n", stdout)
	assert.Empty(t, stderr)
	assert.True(t, runner.closeCalled)
}

func TestInspectDirtyVersion(t *testing.T) {
	runner := &fakeMigration{version: 13, dirty: true}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{"-source", "file:///migrations", "inspect"},
		runner,
	)

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "version=13 dirty=true\n", stdout)
	assert.Empty(t, stderr)
}

func TestInspectDatabaseWithoutVersion(t *testing.T) {
	runner := &fakeMigration{versionErr: migrate.ErrNilVersion}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{"-source", "file:///migrations", "inspect"},
		runner,
	)

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "version=nil dirty=false\n", stdout)
	assert.Empty(t, stderr)
}

func TestForceDirtyVersion(t *testing.T) {
	runner := &fakeMigration{version: 13, dirty: true}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{
			"-source", "file:///migrations",
			"-expected-version", "13",
			"-target-version", "12",
			"-confirm-target-version", "12",
			"force",
		},
		runner,
	)

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "forced version=12 dirty=false\n", stdout)
	assert.Empty(t, stderr)
	assert.True(t, runner.forceCalled)
	assert.Equal(t, 12, runner.forcedVersion)
	assert.True(t, runner.closeCalled)
}

func TestForceCanResetToNilVersion(t *testing.T) {
	runner := &fakeMigration{version: 1, dirty: true}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{
			"-source", "file:///migrations",
			"-expected-version", "1",
			"-target-version", "-1",
			"-confirm-target-version", "-1",
			"force",
		},
		runner,
	)

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "forced version=-1 dirty=false\n", stdout)
	assert.Empty(t, stderr)
	assert.Equal(t, -1, runner.forcedVersion)
}

func TestForceRejectsCleanDatabase(t *testing.T) {
	runner := &fakeMigration{version: 13}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{
			"-source", "file:///migrations",
			"-expected-version", "13",
			"-target-version", "12",
			"-confirm-target-version", "12",
			"force",
		},
		runner,
	)

	assert.Equal(t, 1, exitCode)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "version 13 is not dirty")
	assert.False(t, runner.forceCalled)
}

func TestForceRejectsUnexpectedDirtyVersion(t *testing.T) {
	runner := &fakeMigration{version: 14, dirty: true}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{
			"-source", "file:///migrations",
			"-expected-version", "13",
			"-target-version", "12",
			"-confirm-target-version", "12",
			"force",
		},
		runner,
	)

	assert.Equal(t, 1, exitCode)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "expected dirty version 13, found 14")
	assert.False(t, runner.forceCalled)
}

func TestForceRejectsConfirmationMismatchBeforeInitialization(t *testing.T) {
	var stderr bytes.Buffer
	factoryCalled := false
	exitCode := run(
		[]string{
			"-source", "file:///migrations",
			"-expected-version", "13",
			"-target-version", "12",
			"-confirm-target-version", "11",
			"force",
		},
		&bytes.Buffer{},
		&stderr,
		func(string, string) (migration, error) {
			factoryCalled = true
			return nil, nil
		},
	)

	assert.Equal(t, 2, exitCode)
	assert.False(t, factoryCalled)
	assert.Contains(t, stderr.String(), "confirmation does not match")
}

func TestForceRequiresEveryGuard(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"-source", "file:///migrations",
			"-expected-version", "13",
			"-target-version", "12",
			"force",
		},
		&bytes.Buffer{},
		&stderr,
		func(string, string) (migration, error) {
			return nil, nil
		},
	)

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stderr.String(), "-confirm-target-version is required")
}

func TestForceRejectsTargetBeyondDirtyVersion(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"-source", "file:///migrations",
			"-expected-version", "13",
			"-target-version", "14",
			"-confirm-target-version", "14",
			"force",
		},
		&bytes.Buffer{},
		&stderr,
		func(string, string) (migration, error) {
			return nil, nil
		},
	)

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stderr.String(), "cannot exceed")
}

func TestInspectReportsVersionAndCloseFailures(t *testing.T) {
	runner := &fakeMigration{
		versionErr:       errors.New("cannot inspect"),
		sourceCloseErr:   errors.New("cannot close source"),
		databaseCloseErr: errors.New("cannot close database"),
	}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{"-source", "file:///migrations", "inspect"},
		runner,
	)

	assert.Equal(t, 1, exitCode)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "cannot inspect")
	assert.Contains(t, stderr, "cannot close source")
	assert.Contains(t, stderr, "cannot close database")
}

func TestForceReportsFailureAndCloses(t *testing.T) {
	runner := &fakeMigration{
		version:  13,
		dirty:    true,
		forceErr: errors.New("cannot force"),
	}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{
			"-source", "file:///migrations",
			"-expected-version", "13",
			"-target-version", "12",
			"-confirm-target-version", "12",
			"force",
		},
		runner,
	)

	assert.Equal(t, 1, exitCode)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "cannot force")
	assert.True(t, runner.closeCalled)
}

func TestForceVerifiesUpdatedMetadata(t *testing.T) {
	runner := &fakeMigration{
		version:            13,
		dirty:              true,
		forceDoesNotUpdate: true,
	}
	exitCode, stdout, stderr := runWithFake(
		t,
		[]string{
			"-source", "file:///migrations",
			"-expected-version", "13",
			"-target-version", "12",
			"-confirm-target-version", "12",
			"force",
		},
		runner,
	)

	assert.Equal(t, 1, exitCode)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "force verification failed")
	assert.True(t, runner.closeCalled)
}

func TestRejectsUnsupportedCommandBeforeInitialization(t *testing.T) {
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
	assert.Contains(t, stderr.String(), "expected command: inspect or force")
}
