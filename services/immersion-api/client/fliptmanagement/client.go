package fliptmanagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	common "github.com/tadoku/tadoku/services/common/client/fliptmanagement"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type Config = common.Config

type Client struct {
	client *common.Client
}

func NewClient(config Config) *Client {
	return &Client{client: common.NewClient(config)}
}

func (c *Client) GetNamedUserAccess(ctx context.Context, flagKey domain.FeatureFlagKey, targetUserID uuid.UUID) (domain.FeatureAccessState, error) {
	state, err := c.client.GetNamedUserAccess(ctx, common.FeatureFlagKey(flagKey), targetUserID)
	return legacyState(state), legacyError(err)
}

func (c *Client) SetNamedUserAccess(ctx context.Context, flagKey domain.FeatureFlagKey, targetUserID uuid.UUID, enabled bool) (domain.FeatureAccessState, error) {
	state, err := c.client.SetNamedUserAccess(ctx, common.FeatureFlagKey(flagKey), targetUserID, enabled)
	return legacyState(state), legacyError(err)
}

func legacyState(state common.State) domain.FeatureAccessState {
	return domain.FeatureAccessState{
		Enabled:     state.Enabled,
		Changed:     state.Changed,
		Environment: state.Environment,
		Revision:    state.Revision,
	}
}

func legacyError(err error) error {
	if errors.Is(err, common.ErrUnavailable) {
		return fmt.Errorf("%w: %v", domain.ErrFeatureAccessUnavailable, err)
	}
	return err
}
