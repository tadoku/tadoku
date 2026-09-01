package fliptmanagement

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

const (
	segmentKey         = "release-log-entry-v2-access"
	segmentName        = "Release log entry v2 access"
	segmentDescription = "Kratos UUIDs explicitly granted access to release log entry v2."
	maxUpdateAttempts  = 2
)

type Config struct {
	URL         string
	Environment string
	HTTPClient  *http.Client
}

type Client struct {
	baseURL     *url.URL
	environment string
	httpClient  *http.Client
}

type segmentConstraint struct {
	Type        string `json:"type"`
	Property    string `json:"property"`
	Operator    string `json:"operator"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type segmentPayload struct {
	Type        string              `json:"@type"`
	Key         string              `json:"key"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	MatchType   string              `json:"matchType"`
	Constraints []segmentConstraint `json:"constraints"`
}

type segmentResource struct {
	NamespaceKey string          `json:"namespaceKey"`
	Key          string          `json:"key"`
	Payload      json.RawMessage `json:"payload"`
}

type segmentResponse struct {
	Resource segmentResource `json:"resource"`
	Revision string          `json:"revision"`
}

type parsedSegment struct {
	Resource struct {
		NamespaceKey string
		Key          string
		Payload      segmentPayload
	}
	Revision string
}

func NewClient(cfg Config) *Client {
	baseURL, _ := url.Parse(strings.TrimRight(cfg.URL, "/"))
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: baseURL, environment: cfg.Environment, httpClient: httpClient}
}

func (c *Client) GetNamedUserAccess(ctx context.Context, flagKey domain.FeatureFlagKey, targetUserID uuid.UUID) (domain.FeatureAccessState, error) {
	if err := c.validate(flagKey, targetUserID); err != nil {
		return domain.FeatureAccessState{}, err
	}
	segment, err := c.getSegment(ctx)
	if err != nil {
		return domain.FeatureAccessState{}, err
	}
	members, err := parseMembers(segment.Resource.Payload.Constraints[0].Value)
	if err != nil {
		return domain.FeatureAccessState{}, err
	}
	return c.state(segment, contains(members, targetUserID.String()), false), nil
}

func (c *Client) SetNamedUserAccess(ctx context.Context, flagKey domain.FeatureFlagKey, targetUserID uuid.UUID, enabled bool) (domain.FeatureAccessState, error) {
	if err := c.validate(flagKey, targetUserID); err != nil {
		return domain.FeatureAccessState{}, err
	}

	for attempt := 0; attempt < maxUpdateAttempts; attempt++ {
		segment, err := c.getSegment(ctx)
		if err != nil {
			return domain.FeatureAccessState{}, err
		}
		members, err := parseMembers(segment.Resource.Payload.Constraints[0].Value)
		if err != nil {
			return domain.FeatureAccessState{}, err
		}
		currentlyEnabled := contains(members, targetUserID.String())
		if currentlyEnabled == enabled {
			return c.state(segment, enabled, false), nil
		}

		if enabled {
			members = append(members, targetUserID.String())
		} else {
			members = remove(members, targetUserID.String())
		}
		members = uniqueSorted(members)
		encodedMembers, err := json.Marshal(members)
		if err != nil {
			return domain.FeatureAccessState{}, fmt.Errorf("encode feature access membership: %w", err)
		}
		segment.Resource.Payload.Constraints[0].Value = string(encodedMembers)

		updated, conflict, err := c.putSegment(ctx, segment)
		if err != nil {
			return domain.FeatureAccessState{}, err
		}
		if conflict {
			continue
		}
		return c.state(updated, enabled, true), nil
	}
	return domain.FeatureAccessState{}, fmt.Errorf("%w: update conflicted", domain.ErrFeatureAccessUnavailable)
}

func (c *Client) validate(flagKey domain.FeatureFlagKey, targetUserID uuid.UUID) error {
	if c == nil || c.baseURL == nil || c.baseURL.Scheme == "" || c.baseURL.Host == "" {
		return errors.New("feature access management is not configured")
	}
	if c.environment != "local" && c.environment != "production" {
		return errors.New("feature access environment is invalid")
	}
	if flagKey != domain.FeatureFlagReleaseLogEntryV2 || targetUserID == uuid.Nil {
		return errors.New("feature access request is invalid")
	}
	return nil
}

func (c *Client) getSegment(ctx context.Context) (parsedSegment, error) {
	endpoint := fmt.Sprintf("%s/api/v2/environments/%s/namespaces/default/resources/flipt.core.Segment/%s", c.baseURL.String(), c.environment, segmentKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return parsedSegment{}, errors.New("create feature access request")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return parsedSegment{}, fmt.Errorf("%w: segment request failed", domain.ErrFeatureAccessUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		drain(resp.Body)
		return parsedSegment{}, fmt.Errorf("%w: segment returned status %d", domain.ErrFeatureAccessUnavailable, resp.StatusCode)
	}
	segment, err := decodeSegment(resp.Body)
	if err != nil {
		return parsedSegment{}, fmt.Errorf("%w: %v", domain.ErrFeatureAccessUnavailable, err)
	}
	return segment, nil
}

func (c *Client) putSegment(ctx context.Context, segment parsedSegment) (parsedSegment, bool, error) {
	payload := struct {
		Key      string         `json:"key"`
		Revision string         `json:"revision"`
		Payload  segmentPayload `json:"payload"`
	}{Key: segmentKey, Revision: segment.Revision, Payload: segment.Resource.Payload}
	body, err := json.Marshal(payload)
	if err != nil {
		return parsedSegment{}, false, errors.New("encode feature access update")
	}
	endpoint := fmt.Sprintf("%s/api/v2/environments/%s/namespaces/default/resources", c.baseURL.String(), c.environment)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return parsedSegment{}, false, errors.New("create feature access update")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return parsedSegment{}, false, fmt.Errorf("%w: update request failed", domain.ErrFeatureAccessUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		drain(resp.Body)
		return parsedSegment{}, true, nil
	}
	if resp.StatusCode != http.StatusOK {
		drain(resp.Body)
		return parsedSegment{}, false, fmt.Errorf("%w: update returned status %d", domain.ErrFeatureAccessUnavailable, resp.StatusCode)
	}
	updated, err := decodeSegment(resp.Body)
	if err != nil {
		return parsedSegment{}, false, fmt.Errorf("%w: %v", domain.ErrFeatureAccessUnavailable, err)
	}
	return updated, false, nil
}

func decodeSegment(reader io.Reader) (parsedSegment, error) {
	var segment segmentResponse
	decoder := json.NewDecoder(io.LimitReader(reader, 1<<20))
	if err := decodeOne(decoder, &segment); err != nil {
		return parsedSegment{}, errors.New("unexpected segment schema")
	}
	var payload segmentPayload
	payloadDecoder := json.NewDecoder(bytes.NewReader(segment.Resource.Payload))
	payloadDecoder.DisallowUnknownFields()
	if err := decodeOne(payloadDecoder, &payload); err != nil {
		return parsedSegment{}, errors.New("unexpected segment schema")
	}
	if segment.Resource.NamespaceKey != "default" || segment.Resource.Key != segmentKey ||
		payload.Type != "flipt.core.Segment" || payload.Key != segmentKey ||
		payload.Name != segmentName || payload.Description != segmentDescription ||
		payload.MatchType != "ALL_MATCH_TYPE" || len(payload.Constraints) != 1 {
		return parsedSegment{}, errors.New("unexpected segment schema")
	}
	constraint := payload.Constraints[0]
	if constraint.Type != "ENTITY_ID_COMPARISON_TYPE" || constraint.Property != "entityId" ||
		constraint.Operator != "isoneof" || constraint.Description != "" || !validRevision(segment.Revision) {
		return parsedSegment{}, errors.New("unexpected segment schema")
	}
	if _, err := parseMembers(constraint.Value); err != nil {
		return parsedSegment{}, err
	}
	parsed := parsedSegment{Revision: segment.Revision}
	parsed.Resource.NamespaceKey = segment.Resource.NamespaceKey
	parsed.Resource.Key = segment.Resource.Key
	parsed.Resource.Payload = payload
	return parsed, nil
}

func parseMembers(value string) ([]string, error) {
	var members []string
	decoder := json.NewDecoder(strings.NewReader(value))
	if err := decodeOne(decoder, &members); err != nil || members == nil {
		return nil, errors.New("unexpected segment schema")
	}
	for _, member := range members {
		id, err := uuid.Parse(member)
		if err != nil || id == uuid.Nil || id.String() != member {
			return nil, errors.New("unexpected segment schema")
		}
	}
	return uniqueSorted(members), nil
}

func (c *Client) state(segment parsedSegment, enabled, changed bool) domain.FeatureAccessState {
	return domain.FeatureAccessState{Enabled: enabled, Changed: changed, Environment: c.environment, Revision: segment.Revision}
}

func decodeOne(decoder *json.Decoder, target any) error {
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

func validRevision(revision string) bool {
	if len(revision) != 40 {
		return false
	}
	for _, char := range revision {
		if !strings.ContainsRune("0123456789abcdef", char) {
			return false
		}
	}
	return true
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func remove(values []string, target string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != target {
			result = append(result, value)
		}
	}
	return result
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func drain(reader io.Reader) {
	_, _ = io.Copy(io.Discard, io.LimitReader(reader, 4096))
}
