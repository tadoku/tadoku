package domain

import (
	"fmt"
	"strings"
)

type PermissionAllowlist struct {
	allowed map[string]struct{} // key: namespace + ":" + relation
}

func ParsePermissionAllowlist(csv string) (PermissionAllowlist, error) {
	out := PermissionAllowlist{allowed: map[string]struct{}{}}
	for _, raw := range strings.Split(csv, ",") {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		parts := strings.Split(entry, ":")
		if len(parts) != 2 {
			return PermissionAllowlist{}, fmt.Errorf("invalid allowlist entry %q (want namespace:relation)", entry)
		}
		namespace := strings.TrimSpace(parts[0])
		relation := strings.TrimSpace(parts[1])
		if namespace == "" || relation == "" {
			return PermissionAllowlist{}, fmt.Errorf("invalid allowlist entry %q (empty namespace or relation)", entry)
		}
		out.allowed[namespace+":"+relation] = struct{}{}
	}
	return out, nil
}

func (a PermissionAllowlist) Allows(namespace, relation string) bool {
	_, ok := a.allowed[namespace+":"+relation]
	return ok
}

// RelationshipMutationAllowlist restricts which services can create/delete which relations.
// Entries are keyed by (serviceName, namespace, relation).
type RelationshipMutationAllowlist struct {
	allowed map[string]map[string]struct{} // serviceName -> (namespace:relation) -> {}
}

// ParseRelationshipMutationAllowlist parses comma-separated entries of the form:
// "service:namespace:relation".
func ParseRelationshipMutationAllowlist(csv string) (RelationshipMutationAllowlist, error) {
	out := RelationshipMutationAllowlist{allowed: map[string]map[string]struct{}{}}
	for _, raw := range strings.Split(csv, ",") {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		parts := strings.Split(entry, ":")
		if len(parts) != 3 {
			return RelationshipMutationAllowlist{}, fmt.Errorf("invalid allowlist entry %q (want service:namespace:relation)", entry)
		}
		service := strings.TrimSpace(parts[0])
		namespace := strings.TrimSpace(parts[1])
		relation := strings.TrimSpace(parts[2])
		if service == "" || namespace == "" || relation == "" {
			return RelationshipMutationAllowlist{}, fmt.Errorf("invalid allowlist entry %q (empty service, namespace, or relation)", entry)
		}
		if out.allowed[service] == nil {
			out.allowed[service] = map[string]struct{}{}
		}
		out.allowed[service][namespace+":"+relation] = struct{}{}
	}
	return out, nil
}

func (a RelationshipMutationAllowlist) Allows(serviceName, namespace, relation string) bool {
	rels, ok := a.allowed[serviceName]
	if !ok {
		return false
	}
	_, ok = rels[namespace+":"+relation]
	return ok
}
