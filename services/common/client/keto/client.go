package keto

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	keto "github.com/ory/keto-client-go"
)

// Subject represents the subject of a permission check or relation.
// For direct subjects, only Namespace and Object are needed.
// For subject sets (e.g., "members of Group:admins"), Relation is also used.
type Subject struct {
	Namespace string
	Object    string
	Relation  string
}

// PermissionChecker provides methods for checking permissions.
type PermissionChecker interface {
	CheckPermission(ctx context.Context, namespace, object, relation string, subject Subject) (bool, error)
	CheckPermissions(ctx context.Context, checks []PermissionCheck) []PermissionResult
}

// RelationManager provides methods for managing relation tuples.
type RelationManager interface {
	AddRelation(ctx context.Context, namespace, object, relation string, subject Subject) error
	DeleteRelation(ctx context.Context, namespace, object, relation string, subject Subject) error
}

// Client implements PermissionChecker and RelationManager.
type Client struct {
	readClient  *keto.APIClient
	writeClient *keto.APIClient
}

// Compile-time interface compliance checks.
var (
	_ PermissionChecker = (*Client)(nil)
	_ RelationManager   = (*Client)(nil)
)

func NewClient(readURL, writeURL string) *Client {
	readCfg := keto.NewConfiguration()
	readCfg.Servers = keto.ServerConfigurations{{URL: readURL}}

	writeCfg := keto.NewConfiguration()
	writeCfg.Servers = keto.ServerConfigurations{{URL: writeURL}}

	return &Client{
		readClient:  keto.NewAPIClient(readCfg),
		writeClient: keto.NewAPIClient(writeCfg),
	}
}

// CheckPermission checks if a subject has a relation on an object.
func (c *Client) CheckPermission(ctx context.Context, namespace, object, relation string, subject Subject) (bool, error) {
	req := c.readClient.PermissionApi.CheckPermission(ctx).
		Namespace(namespace).
		Object(object).
		Relation(relation).
		SubjectSetNamespace(subject.Namespace).
		SubjectSetObject(subject.Object).
		SubjectSetRelation(subject.Relation)

	result, res, err := c.readClient.PermissionApi.CheckPermissionExecute(req)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusForbidden {
			return false, nil
		}
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	return result.GetAllowed(), nil
}

// AddRelation creates a relation tuple in Keto.
func (c *Client) AddRelation(ctx context.Context, namespace, object, relation string, subject Subject) error {
	body := keto.CreateRelationshipBody{
		Namespace: &namespace,
		Object:    &object,
		Relation:  &relation,
		SubjectSet: &keto.SubjectSet{
			Namespace: subject.Namespace,
			Object:    subject.Object,
			Relation:  subject.Relation,
		},
	}

	req := c.writeClient.RelationshipApi.CreateRelationship(ctx).CreateRelationshipBody(body)
	_, _, err := c.writeClient.RelationshipApi.CreateRelationshipExecute(req)
	if err != nil {
		return fmt.Errorf("failed to create relation: %w", err)
	}

	return nil
}

// DeleteRelation removes a relation tuple from Keto.
func (c *Client) DeleteRelation(ctx context.Context, namespace, object, relation string, subject Subject) error {
	req := c.writeClient.RelationshipApi.DeleteRelationships(ctx).
		Namespace(namespace).
		Object(object).
		Relation(relation).
		SubjectSetNamespace(subject.Namespace).
		SubjectSetObject(subject.Object).
		SubjectSetRelation(subject.Relation)

	_, err := c.writeClient.RelationshipApi.DeleteRelationshipsExecute(req)
	if err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}

	return nil
}

// PermissionCheck represents a single permission check request.
type PermissionCheck struct {
	Namespace string
	Object    string
	Relation  string
	Subject   Subject
}

// PermissionResult represents the result of a single permission check.
type PermissionResult struct {
	Check   PermissionCheck
	Allowed bool
	Err     error
}

// DefaultMaxConcurrency is the default maximum number of concurrent permission checks.
const DefaultMaxConcurrency = 10

// CheckPermissions checks multiple permissions in parallel.
// Returns results in the same order as the input checks.
// Limits concurrency to DefaultMaxConcurrency to avoid overwhelming the server.
func (c *Client) CheckPermissions(ctx context.Context, checks []PermissionCheck) []PermissionResult {
	results := make([]PermissionResult, len(checks))
	if len(checks) == 0 {
		return results
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, DefaultMaxConcurrency)

	for i, check := range checks {
		wg.Add(1)
		go func(i int, check PermissionCheck) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			allowed, err := c.CheckPermission(ctx, check.Namespace, check.Object, check.Relation, check.Subject)
			results[i] = PermissionResult{
				Check:   check,
				Allowed: allowed,
				Err:     err,
			}
		}(i, check)
	}

	wg.Wait()
	return results
}
