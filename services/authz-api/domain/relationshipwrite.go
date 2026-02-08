package domain

import (
	"context"
	"fmt"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
)

type RelationshipWriteRequest struct {
	Namespace string
	Object    string
	Relation  string
	Subject   ketoclient.Subject
}

type RelationshipWriter struct {
	keto      ketoclient.AuthorizationClient
	allowlist RelationshipMutationAllowlist
}

func NewRelationshipWriter(keto ketoclient.AuthorizationClient, allowlist RelationshipMutationAllowlist) *RelationshipWriter {
	return &RelationshipWriter{keto: keto, allowlist: allowlist}
}

func (s *RelationshipWriter) Create(ctx context.Context, callerService string, req RelationshipWriteRequest) error {
	if req.Namespace == "" || req.Object == "" || req.Relation == "" {
		return fmt.Errorf("%w: namespace, object, and relation are required", commondomain.ErrRequestInvalid)
	}
	if !s.allowlist.Allows(callerService, req.Namespace, req.Relation) {
		return commondomain.ErrForbidden
	}
	if err := s.keto.AddRelation(ctx, req.Namespace, req.Object, req.Relation, req.Subject); err != nil {
		return commondomain.ErrAuthzUnavailable
	}
	return nil
}

func (s *RelationshipWriter) Delete(ctx context.Context, callerService string, req RelationshipWriteRequest) error {
	if req.Namespace == "" || req.Object == "" || req.Relation == "" {
		return fmt.Errorf("%w: namespace, object, and relation are required", commondomain.ErrRequestInvalid)
	}
	if !s.allowlist.Allows(callerService, req.Namespace, req.Relation) {
		return commondomain.ErrForbidden
	}
	if err := s.keto.DeleteRelation(ctx, req.Namespace, req.Object, req.Relation, req.Subject); err != nil {
		return commondomain.ErrAuthzUnavailable
	}
	return nil
}

