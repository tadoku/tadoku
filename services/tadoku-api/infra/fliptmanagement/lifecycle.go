package fliptmanagement

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func (c *Client) ProvisionTestNamespace(ctx context.Context, key tenant.TestKey, resources []Resource) error {
	target, err := c.testNamespaceTarget(ctx, key)
	if err != nil {
		return err
	}
	collection := fmt.Sprintf("%s/api/v2/environments/%s/namespaces",
		c.baseURL.String(), url.PathEscape(target.Environment))
	endpoint := collection + "/" + url.PathEscape(target.Namespace)
	for _, resource := range resources {
		if resource.Key == "" || !json.Valid(resource.Payload) ||
			(resource.resourceType != "flipt.core.Segment" && resource.resourceType != "flipt.core.Flag") {
			return errors.New("Flipt lifecycle resources must come from ParseFeatures")
		}
	}

	status, err := c.lifecycleRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	if status == http.StatusNotFound {
		payload := struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		}{Key: target.Namespace, Name: key.String()}
		encoded, _ := json.Marshal(payload)
		status, err = c.lifecycleRequest(ctx, http.MethodPost, collection, encoded)
		if err != nil {
			return err
		}
		if status != http.StatusOK && status != http.StatusCreated && status != http.StatusConflict {
			return lifecycleStatus("namespace create", status)
		}
	} else if status != http.StatusOK {
		return lifecycleStatus("namespace read", status)
	}

	for _, resource := range resources {
		resourceURL := endpoint + "/resources/" + url.PathEscape(resource.resourceType) + "/" +
			url.PathEscape(resource.Key)
		status, err := c.lifecycleRequest(ctx, http.MethodGet, resourceURL, nil)
		if err != nil {
			return err
		}
		if status == http.StatusOK {
			continue
		}
		if status != http.StatusNotFound {
			return lifecycleStatus("resource read", status)
		}
		encoded, err := json.Marshal(resource)
		if err != nil {
			return fmt.Errorf("encode Flipt lifecycle resource: %w", err)
		}
		status, err = c.lifecycleRequest(ctx, http.MethodPost, endpoint+"/resources", encoded)
		if err != nil {
			return err
		}
		if status != http.StatusOK && status != http.StatusCreated && status != http.StatusConflict {
			return lifecycleStatus("resource create", status)
		}
	}
	return nil
}

func (c *Client) DeleteTestNamespace(ctx context.Context, key tenant.TestKey) error {
	target, err := c.testNamespaceTarget(ctx, key)
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("%s/api/v2/environments/%s/namespaces/%s",
		c.baseURL.String(), url.PathEscape(target.Environment), url.PathEscape(target.Namespace))
	status, err := c.lifecycleRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent && status != http.StatusNotFound {
		return lifecycleStatus("namespace delete", status)
	}
	return nil
}

func (c *Client) testNamespaceTarget(ctx context.Context, key tenant.TestKey) (flipt.Target, error) {
	if c == nil || c.baseURL == nil || c.baseURL.Scheme == "" || c.baseURL.Host == "" || key.String() == "" {
		return flipt.Target{}, errors.New("Flipt lifecycle requires a configured provider and parsed test tenant")
	}
	return c.targets.Resolve(tenant.WithKey(ctx, key.Key()))
}

func (c *Client) lifecycleRequest(ctx context.Context, method, endpoint string, body []byte) (int, error) {
	request, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create Flipt lifecycle request: %w", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return 0, fmt.Errorf("%w: Flipt lifecycle request failed", ErrUnavailable)
	}
	defer response.Body.Close()
	drain(response.Body)
	return response.StatusCode, nil
}

func lifecycleStatus(operation string, status int) error {
	return fmt.Errorf("%w: Flipt %s returned status %d", ErrUnavailable, operation, status)
}
