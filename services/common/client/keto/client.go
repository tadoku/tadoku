package keto

import (
	"context"
	"fmt"
	"net/http"

	keto "github.com/ory/keto-client-go"
	"golang.org/x/sync/errgroup"
)

// Keto supports two encodings:
// - Direct subject: subject_id
// - Subject set: subject_set (namespace, object, relation)
//
// Exactly one of ID or Set must be provided.
type Subject struct {
	// ID is sent as subject_id.
	ID string
	// Set is sent as subject_set.*.
	Set *SubjectSet
}

type SubjectSet struct {
	Namespace string
	Object    string
	Relation  string
}

type AuthorizationReader interface {
	CheckPermission(ctx context.Context, namespace, object, relation string, subject Subject) (bool, error)
	CheckPermissions(ctx context.Context, checks []PermissionCheck) []PermissionResult

	// ListSubjectIDsForRelation returns all direct subject IDs which have
	// (namespace, object, relation). Subject sets are ignored.
	ListSubjectIDsForRelation(ctx context.Context, namespace, object, relation string) ([]string, error)
}

type AuthorizationClient interface {
	AuthorizationReader
	AddRelation(ctx context.Context, namespace, object, relation string, subject Subject) error
	DeleteRelation(ctx context.Context, namespace, object, relation string, subject Subject) error
}

type Client struct {
	readClient  *keto.APIClient
	writeClient *keto.APIClient
}

var (
	_ AuthorizationReader = (*Client)(nil)
	_ AuthorizationClient = (*Client)(nil)
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

// NewReadClient creates a client that can only check permissions.
// Relation operations will return an error.
func NewReadClient(readURL string) AuthorizationReader {
	readCfg := keto.NewConfiguration()
	readCfg.Servers = keto.ServerConfigurations{{URL: readURL}}

	return &Client{
		readClient: keto.NewAPIClient(readCfg),
	}
}

// Keto returns HTTP 403 when permission is denied, which is treated as (false, nil).
func (c *Client) CheckPermission(ctx context.Context, namespace, object, relation string, subject Subject) (bool, error) {
	req := c.readClient.PermissionApi.CheckPermission(ctx).
		Namespace(namespace).
		Object(object).
		Relation(relation)

	switch {
	case subject.ID != "":
		req = req.SubjectId(subject.ID)
	case subject.Set != nil:
		req = req.
			SubjectSetNamespace(subject.Set.Namespace).
			SubjectSetObject(subject.Set.Object).
			SubjectSetRelation(subject.Set.Relation)
	default:
		return false, fmt.Errorf("subject must set either ID or Set")
	}

	result, res, err := c.readClient.PermissionApi.CheckPermissionExecute(req)
	if err != nil {
		// Keto returns 403 when permission is denied (not an error condition)
		if res != nil && res.StatusCode == http.StatusForbidden {
			return false, nil
		}
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	return result.GetAllowed(), nil
}

const defaultRelationshipPageSize int64 = 500

func (c *Client) ListSubjectIDsForRelation(ctx context.Context, namespace, object, relation string) ([]string, error) {
	req := c.readClient.RelationshipApi.GetRelationships(ctx).
		Namespace(namespace).
		Object(object).
		Relation(relation).
		PageSize(defaultRelationshipPageSize)

	subjectIDs := make([]string, 0, 16)
	pageToken := ""
	for {
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		rels, _, err := c.readClient.RelationshipApi.GetRelationshipsExecute(req)
		if err != nil {
			return nil, fmt.Errorf("failed to list relationships: %w", err)
		}

		for _, t := range rels.GetRelationTuples() {
			if t.SubjectId != nil && *t.SubjectId != "" {
				subjectIDs = append(subjectIDs, *t.SubjectId)
			}
		}

		pageToken = rels.GetNextPageToken()
		if pageToken == "" {
			return subjectIDs, nil
		}
	}
}

func (c *Client) AddRelation(ctx context.Context, namespace, object, relation string, subject Subject) error {
	if c.writeClient == nil {
		return fmt.Errorf("keto write client not configured")
	}

	body := keto.CreateRelationshipBody{
		Namespace: &namespace,
		Object:    &object,
		Relation:  &relation,
	}

	switch {
	case subject.ID != "":
		body.SubjectId = &subject.ID
	case subject.Set != nil:
		body.SubjectSet = &keto.SubjectSet{
			Namespace: subject.Set.Namespace,
			Object:    subject.Set.Object,
			Relation:  subject.Set.Relation,
		}
	default:
		return fmt.Errorf("subject must set either ID or Set")
	}

	req := c.writeClient.RelationshipApi.CreateRelationship(ctx).CreateRelationshipBody(body)
	_, res, err := c.writeClient.RelationshipApi.CreateRelationshipExecute(req)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusConflict {
			return nil
		}
		return fmt.Errorf("failed to create relation: %w", err)
	}

	return nil
}

func (c *Client) DeleteRelation(ctx context.Context, namespace, object, relation string, subject Subject) error {
	if c.writeClient == nil {
		return fmt.Errorf("keto write client not configured")
	}

	req := c.writeClient.RelationshipApi.DeleteRelationships(ctx).
		Namespace(namespace).
		Object(object).
		Relation(relation)

	switch {
	case subject.ID != "":
		req = req.SubjectId(subject.ID)
	case subject.Set != nil:
		req = req.
			SubjectSetNamespace(subject.Set.Namespace).
			SubjectSetObject(subject.Set.Object).
			SubjectSetRelation(subject.Set.Relation)
	default:
		return fmt.Errorf("subject must set either ID or Set")
	}

	res, err := c.writeClient.RelationshipApi.DeleteRelationshipsExecute(req)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete relation: %w", err)
	}

	return nil
}

type PermissionCheck struct {
	Namespace string
	Object    string
	Relation  string
	Subject   Subject
}

type PermissionResult struct {
	Check   PermissionCheck
	Allowed bool
	Err     error
}

const defaultMaxConcurrency = 10

// CheckPermissions checks multiple permissions in parallel and returns results
// in input order. Concurrency is capped to avoid overwhelming Keto.
// Cancelled checks record ctx.Err() instead of failing the batch.
func (c *Client) CheckPermissions(ctx context.Context, checks []PermissionCheck) []PermissionResult {
	results := make([]PermissionResult, len(checks))
	if len(checks) == 0 {
		return results
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(defaultMaxConcurrency)

	for i, check := range checks {
		i, check := i, check
		g.Go(func() error {
			allowed, err := c.CheckPermission(ctx, check.Namespace, check.Object, check.Relation, check.Subject)
			results[i] = PermissionResult{
				Check:   check,
				Allowed: allowed,
				Err:     err,
			}
			return nil // Don't fail-fast; collect all results
		})
	}

	g.Wait()
	return results
}
