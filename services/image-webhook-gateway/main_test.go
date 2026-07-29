package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("GATEWAY_ADDRESS", "test-address")
	t.Setenv("GATEWAY_IMAGE_UPDATER_WEBHOOK_URL", "https://updater.example/webhook")
	t.Setenv("GATEWAY_IMAGE_UPDATER_NAMESPACE", "test-namespace")
	t.Setenv("GATEWAY_IMAGE_UPDATER_NAME", "test-updater")
	t.Setenv("GATEWAY_QUEUE_SIZE", "7")
	t.Setenv("GATEWAY_GHCR_WEBHOOK_SECRET", "test-secret")

	cfg, err := loadConfig()

	require.NoError(t, err)
	assert.Equal(t, "test-address", cfg.Address)
	assert.Equal(t, "https://updater.example/webhook", cfg.ImageUpdaterWebhookURL)
	assert.Equal(t, "test-namespace", cfg.ImageUpdaterNamespace)
	assert.Equal(t, "test-updater", cfg.ImageUpdaterName)
	assert.Equal(t, 7, cfg.QueueSize)
	assert.Equal(t, "test-secret", cfg.GHCRWebhookSecret)
}
