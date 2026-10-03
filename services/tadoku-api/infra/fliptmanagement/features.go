package fliptmanagement

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

type Resource struct {
	Key          string          `json:"key"`
	Payload      json.RawMessage `json:"payload"`
	resourceType string
}

type featuresFile struct {
	Version   string        `yaml:"version"`
	Namespace seedNamespace `yaml:"namespace"`
	Flags     []seedFlag    `yaml:"flags"`
	Segments  []seedSegment `yaml:"segments"`
}

type seedNamespace struct {
	Key         string `yaml:"key"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type seedFlag struct {
	Key         string        `yaml:"key"`
	Name        string        `yaml:"name"`
	Type        string        `yaml:"type"`
	Description string        `yaml:"description"`
	Enabled     bool          `yaml:"enabled"`
	Rollouts    []seedRollout `yaml:"rollouts"`
}

type seedRollout struct {
	Description string              `yaml:"description"`
	Segment     *seedSegmentRollout `yaml:"segment"`
	Threshold   *thresholdRollout   `yaml:"threshold"`
}

type seedSegmentRollout struct {
	Keys     []string `yaml:"keys"`
	Operator string   `yaml:"operator"`
	Value    bool     `yaml:"value"`
}

type thresholdRollout struct {
	Percentage float64 `yaml:"percentage" json:"percentage"`
	Value      bool    `yaml:"value" json:"value"`
}

type seedSegment struct {
	Key         string           `yaml:"key"`
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	MatchType   string           `yaml:"match_type"`
	Constraints []seedConstraint `yaml:"constraints"`
}

type seedConstraint struct {
	Type        string    `yaml:"type"`
	Property    string    `yaml:"property"`
	Operator    string    `yaml:"operator"`
	Value       yaml.Node `yaml:"value"`
	Description string    `yaml:"description"`
}

type flagPayload struct {
	Type        string           `json:"@type"`
	Key         string           `json:"key"`
	Name        string           `json:"name"`
	FlagType    string           `json:"type"`
	Description string           `json:"description"`
	Enabled     bool             `json:"enabled"`
	Rollouts    []rolloutPayload `json:"rollouts"`
}

type rolloutPayload struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Segment     *segmentRolloutPayload `json:"segment,omitempty"`
	Threshold   *thresholdRollout      `json:"threshold,omitempty"`
}

type segmentRolloutPayload struct {
	Segments []string `json:"segments"`
	Operator string   `json:"segmentOperator"`
	Value    bool     `json:"value"`
}

func ParseFeatures(reader io.Reader) ([]Resource, error) {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)
	var file featuresFile
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("parse Flipt features: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, errors.New("Flipt features must contain exactly one YAML document")
	}
	if file.Version != "1.6" || file.Namespace.Key == "" {
		return nil, errors.New("Flipt features require version 1.6 and a namespace key")
	}

	resources := make([]Resource, 0, len(file.Segments)+len(file.Flags))
	seen := make(map[string]bool)
	appendResource := func(resourceType, key string, payload any) error {
		identity := resourceType + "/" + key
		if key == "" || seen[identity] {
			return errors.New("Flipt features contain an empty or duplicate resource key")
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode Flipt resource: %w", err)
		}
		seen[identity] = true
		resources = append(resources, Resource{Key: key, Payload: encoded, resourceType: resourceType})
		return nil
	}

	for _, segment := range file.Segments {
		payload, err := convertSegment(segment)
		if err != nil {
			return nil, err
		}
		if err := appendResource(payload.Type, segment.Key, payload); err != nil {
			return nil, err
		}
	}
	for _, flag := range file.Flags {
		payload, err := convertFlag(flag)
		if err != nil {
			return nil, err
		}
		if err := appendResource(payload.Type, flag.Key, payload); err != nil {
			return nil, err
		}
	}
	return resources, nil
}

func convertSegment(segment seedSegment) (segmentPayload, error) {
	if segment.MatchType == "" {
		segment.MatchType = "ALL_MATCH_TYPE"
	}
	if segment.MatchType != "ALL_MATCH_TYPE" && segment.MatchType != "ANY_MATCH_TYPE" {
		return segmentPayload{}, errors.New("Flipt segment has an invalid match type")
	}
	payload := segmentPayload{
		Type:        "flipt.core.Segment",
		Key:         segment.Key,
		Name:        segment.Name,
		Description: segment.Description,
		MatchType:   segment.MatchType,
		Constraints: make([]segmentConstraint, 0, len(segment.Constraints)),
	}
	for _, constraint := range segment.Constraints {
		value, err := constraintValue(constraint.Value)
		if err != nil {
			return segmentPayload{}, err
		}
		payload.Constraints = append(payload.Constraints, segmentConstraint{
			Type:        constraint.Type,
			Property:    constraint.Property,
			Operator:    constraint.Operator,
			Value:       value,
			Description: constraint.Description,
		})
	}
	return payload, nil
}

func constraintValue(node yaml.Node) (string, error) {
	if node.Kind != yaml.ScalarNode && node.Kind != yaml.SequenceNode {
		return "", errors.New("Flipt constraint value must be a scalar or list")
	}
	if node.Kind == yaml.ScalarNode {
		if node.Tag == "!!null" {
			return "", errors.New("Flipt constraint value must not be null")
		}
		return node.Value, nil
	}

	var values []any
	if err := node.Decode(&values); err != nil {
		return "", fmt.Errorf("parse Flipt constraint values: %w", err)
	}
	for _, value := range values {
		switch value.(type) {
		case string, int, uint64, float64, bool:
		default:
			return "", errors.New("Flipt constraint list must contain scalar values")
		}
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("encode Flipt constraint values: %w", err)
	}
	return string(encoded), nil
}

func convertFlag(flag seedFlag) (flagPayload, error) {
	if flag.Type != "BOOLEAN_FLAG_TYPE" {
		return flagPayload{}, errors.New("Flipt features require boolean flags")
	}
	payload := flagPayload{
		Type:        "flipt.core.Flag",
		Key:         flag.Key,
		Name:        flag.Name,
		FlagType:    flag.Type,
		Description: flag.Description,
		Enabled:     flag.Enabled,
		Rollouts:    make([]rolloutPayload, 0, len(flag.Rollouts)),
	}
	for _, rollout := range flag.Rollouts {
		converted := rolloutPayload{Description: rollout.Description}
		switch {
		case rollout.Segment != nil && rollout.Threshold == nil:
			segment := rollout.Segment
			if segment.Operator == "" {
				segment.Operator = "OR_SEGMENT_OPERATOR"
			}
			if len(segment.Keys) == 0 || (segment.Operator != "OR_SEGMENT_OPERATOR" &&
				segment.Operator != "AND_SEGMENT_OPERATOR") {
				return flagPayload{}, errors.New("Flipt segment rollout requires keys and a valid operator")
			}
			converted.Type = "SEGMENT_ROLLOUT_TYPE"
			converted.Segment = &segmentRolloutPayload{
				Segments: segment.Keys,
				Operator: segment.Operator,
				Value:    segment.Value,
			}
		case rollout.Threshold != nil && rollout.Segment == nil:
			if rollout.Threshold.Percentage < 0 || rollout.Threshold.Percentage > 100 {
				return flagPayload{}, errors.New("Flipt threshold rollout percentage must be between 0 and 100")
			}
			converted.Type = "THRESHOLD_ROLLOUT_TYPE"
			converted.Threshold = rollout.Threshold
		default:
			return flagPayload{}, errors.New("Flipt rollout must contain exactly one segment or threshold")
		}
		payload.Rollouts = append(payload.Rollouts, converted)
	}
	return payload, nil
}
