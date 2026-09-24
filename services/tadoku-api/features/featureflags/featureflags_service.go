package featureflags

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	commonflags "github.com/tadoku/tadoku/services/common/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
)

type Service struct {
	evaluator *commonflags.Evaluator
	client    *fliptmanagement.Client
}

func NewService(evaluator *commonflags.Evaluator, client *fliptmanagement.Client) *Service {
	return &Service{evaluator: evaluator, client: client}
}

func (s *Service) EvaluatePublic(ctx context.Context, subject string) PublicDecisions {
	decisions := s.evaluator.EvaluatePublicForSubject(ctx, subject)
	return PublicDecisions{ReleaseLogEntryV2: decisions.ReleaseLogEntryV2}
}

func (s *Service) ValidateRequest(flagKey string, targetUserID uuid.UUID) error {
	_, err := validate(flagKey, targetUserID)
	return err
}

func (s *Service) GetNamedUserAccess(ctx context.Context, flagKey string, targetUserID uuid.UUID) (State, error) {
	segment, err := validate(flagKey, targetUserID)
	if err != nil {
		return State{}, err
	}

	state, err := s.client.GetNamedUserAccess(ctx, segment, targetUserID)
	if err != nil {
		if errors.Is(err, fliptmanagement.ErrUnavailable) {
			return State{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
		return State{}, fmt.Errorf("fetch feature access: %w", err)
	}
	return fromClient(state), nil
}

func (s *Service) Grant(ctx context.Context, flagKey string, targetUserID uuid.UUID) (State, error) {
	return s.setNamedUserAccess(ctx, flagKey, targetUserID, true)
}

func (s *Service) Revoke(ctx context.Context, flagKey string, targetUserID uuid.UUID) (State, error) {
	return s.setNamedUserAccess(ctx, flagKey, targetUserID, false)
}

func (s *Service) setNamedUserAccess(ctx context.Context, flagKey string, targetUserID uuid.UUID, enabled bool) (State, error) {
	segment, err := validate(flagKey, targetUserID)
	if err != nil {
		return State{}, err
	}

	state, err := s.client.SetNamedUserAccess(ctx, segment, targetUserID, enabled)
	if err != nil {
		if errors.Is(err, fliptmanagement.ErrUnavailable) {
			return State{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
		return State{}, fmt.Errorf("update feature access: %w", err)
	}
	return fromClient(state), nil
}

func fromClient(state fliptmanagement.State) State {
	return State{
		Enabled:     state.Enabled,
		Changed:     state.Changed,
		Environment: state.Environment,
		Revision:    state.Revision,
	}
}
