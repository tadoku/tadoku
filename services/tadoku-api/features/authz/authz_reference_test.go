package authz

import (
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

// refRule is a test-only copy of a permission rule. The reference model compiles it into the rules the checker
// evaluates and interprets it itself, so the reference evaluator shares no code with the checker's evaluator.
type refRule struct {
	kind       string
	name       string
	permit     string
	objectType permissions.ObjectType
	objectID   string
	rules      []refRule
}

func refRelated(relation string) refRule { return refRule{kind: "related", name: relation} }

func refAllowed(permit string) refRule { return refRule{kind: "allowed", permit: permit} }

func refVia(relation, permit string) refRule {
	return refRule{kind: "via", name: relation, permit: permit}
}

func refPermitOn(objectType permissions.ObjectType, objectID, permit string) refRule {
	return refRule{kind: "permitOn", objectType: objectType, objectID: objectID, permit: permit}
}

func refAnyOf(rules ...refRule) refRule { return refRule{kind: "anyOf", rules: rules} }

func refAllOf(rules ...refRule) refRule { return refRule{kind: "allOf", rules: rules} }

func refExcept(base, subtract refRule) refRule {
	return refRule{kind: "except", rules: []refRule{base, subtract}}
}

func (r refRule) compile() permissions.Rule {
	rules := make([]permissions.Rule, len(r.rules))
	for i, rule := range r.rules {
		rules[i] = rule.compile()
	}
	switch r.kind {
	case "related":
		return permissions.Related(r.name)
	case "allowed":
		return permissions.Allowed(r.permit)
	case "via":
		return permissions.Via(r.name, r.permit)
	case "permitOn":
		return permissions.PermitOn(r.objectType, r.objectID, r.permit)
	case "anyOf":
		return permissions.AnyOf(rules...)
	case "allOf":
		return permissions.AllOf(rules...)
	default:
		return permissions.Except(rules[0], rules[1])
	}
}

var groupMembers, groupMemberPermit = permissions.SetOf("group", "members"), permissions.SetOf("group", "member")

// refNamespace is one namespace of the reference model. Together they cover every rule kind: Related over users and
// over subject sets of a relation and of a permit, Allowed, Via over contest links and group parents, PermitOn, AnyOf,
// AllOf and Except. Nested groups and parent links can form cycles, so random graphs also reach the depth limit.
type refNamespace struct {
	relations permissions.Relations
	permits   map[string]refRule
}

var refNamespaces = map[permissions.ObjectType]refNamespace{
	"site": {
		relations: permissions.Relations{"admins": {permissions.Users}},
		permits:   map[string]refRule{"admin": refRelated("admins")},
	},
	"group": {
		relations: permissions.Relations{
			"members": {permissions.Users, permissions.SetOf("group", "members")},
			"parent":  {permissions.ObjectOf("group")},
		},
		permits: map[string]refRule{"member": refAnyOf(refRelated("members"), refVia("parent", "member"))},
	},
	"contest": {
		relations: permissions.Relations{
			"owners":     {permissions.Users},
			"moderators": {permissions.Users, groupMembers, groupMemberPermit},
			"suspended":  {permissions.Users, groupMemberPermit},
			"verified":   {permissions.Users},
		},
		permits: map[string]refRule{
			"edit":     refAnyOf(refRelated("owners"), refPermitOn("site", "root", "admin")),
			"moderate": refExcept(refAnyOf(refAllowed("edit"), refRelated("moderators")), refRelated("suspended")),
			"review":   refAllOf(refAllowed("moderate"), refRelated("verified")),
		},
	},
	"log": {
		relations: permissions.Relations{
			"owner":   {permissions.Users},
			"contest": {permissions.ObjectOf("contest")},
		},
		permits: map[string]refRule{
			"edit": refAnyOf(refRelated("owner"), refVia("contest", "moderate")),
			"hide": refAnyOf(refPermitOn("site", "root", "admin"), refVia("contest", "review")),
		},
	},
}

func defineReference(objectType permissions.ObjectType) permissions.Namespace {
	namespace := refNamespaces[objectType]
	permits := permissions.Permits{}
	for name, rule := range namespace.permits {
		permits[name] = rule.compile()
	}
	return permissions.MustDefine(objectType, namespace.relations, permits)
}

var (
	refSites    = defineReference("site")
	refGroups   = defineReference("group")
	refContests = defineReference("contest")
	refLogs     = defineReference("log")
	refModel    = mustModel(refSites, refGroups, refContests, refLogs)

	refByType = map[permissions.ObjectType]permissions.Namespace{
		"site":    refSites,
		"group":   refGroups,
		"contest": refContests,
		"log":     refLogs,
	}
)

type refVerdict int

const (
	refDenied refVerdict = iota
	refGranted
	refFailed
)

func (v refVerdict) String() string {
	if v < refDenied || v > refFailed {
		return "another error"
	}
	return [...]string{"denied", "granted", "depth error"}[v]
}

// reference is the naive evaluator the properties compare the checker with: recursive, unbatched and in memory over
// the generated rows. A rule decides its result when it can: AnyOf passes when any rule passes and AllOf and Except
// deny when a rule denies them, otherwise a failed rule fails it. Each Related, Allowed, PermitOn and Via step and
// each subject-set hop adds one to the depth, and a permit, relation or link read past depth 8 fails.
type reference struct {
	rows    []permissions.Relationship
	answers map[refKey]refVerdict
}

type refKey struct {
	object permissions.Object
	name   string
	user   uuid.UUID
	depth  int
}

func newReference(rows []permissions.Relationship) *reference {
	return &reference{rows: rows, answers: map[refKey]refVerdict{}}
}

func (ref *reference) check(object permissions.Object, permit string, user uuid.UUID) refVerdict {
	return ref.permit(object, permit, user, 0)
}

func (ref *reference) permit(object permissions.Object, permit string, user uuid.UUID, depth int) refVerdict {
	key := refKey{object: object, name: permit, user: user, depth: depth}
	if verdict, ok := ref.answers[key]; ok {
		return verdict
	}

	verdict := refFailed
	if rule, ok := refNamespaces[object.Type()].permits[permit]; ok && depth <= 8 {
		verdict = ref.rule(object, rule, user, depth)
	}
	ref.answers[key] = verdict
	return verdict
}

func (ref *reference) rule(object permissions.Object, rule refRule, user uuid.UUID, depth int) refVerdict {
	var children []refVerdict
	for _, child := range rule.rules {
		children = append(children, ref.rule(object, child, user, depth))
	}

	switch rule.kind {
	case "related":
		return ref.relation(object, rule.name, user, depth+1)
	case "allowed":
		return ref.permit(object, rule.permit, user, depth+1)
	case "permitOn":
		return ref.permit(refByType[rule.objectType].Object(rule.objectID), rule.permit, user, depth+1)
	case "via":
		return ref.via(object, rule.name, rule.permit, user, depth+1)
	case "anyOf":
		return refAny(children)
	case "allOf":
		return refAll(children)
	default:
		return refAll([]refVerdict{children[0], refNot(children[1])})
	}
}

func (ref *reference) relation(object permissions.Object, relation string, user uuid.UUID, depth int) refVerdict {
	key := refKey{object: object, name: relation, user: user, depth: depth}
	if verdict, ok := ref.answers[key]; ok {
		return verdict
	}

	verdict := refFailed
	if depth <= 8 {
		var sets []refVerdict
		for _, row := range ref.rows {
			if row.Object() != object || row.Relation() != relation {
				continue
			}
			if subject, ok := row.User(); ok && subject == user {
				sets = append(sets, refGranted)
			}
			if set, setRelation, ok := row.Set(); ok {
				if _, isPermit := refNamespaces[set.Type()].permits[setRelation]; isPermit {
					sets = append(sets, ref.permit(set, setRelation, user, depth+1))
				} else {
					sets = append(sets, ref.relation(set, setRelation, user, depth+1))
				}
			}
		}
		verdict = refAny(sets)
	}
	ref.answers[key] = verdict
	return verdict
}

func (ref *reference) via(object permissions.Object, relation, permit string, user uuid.UUID, depth int) refVerdict {
	if depth > 8 {
		return refFailed
	}

	var targets []refVerdict
	for _, row := range ref.rows {
		if target, ok := row.Target(); ok && row.Object() == object && row.Relation() == relation {
			targets = append(targets, ref.permit(target, permit, user, depth))
		}
	}
	return refAny(targets)
}

func refAny(verdicts []refVerdict) refVerdict {
	result := refDenied
	for _, verdict := range verdicts {
		if verdict == refGranted {
			return refGranted
		}
		if verdict == refFailed {
			result = refFailed
		}
	}
	return result
}

func refAll(verdicts []refVerdict) refVerdict {
	result := refGranted
	for _, verdict := range verdicts {
		if verdict == refDenied {
			return refDenied
		}
		if verdict == refFailed {
			result = refFailed
		}
	}
	return result
}

func refNot(verdict refVerdict) refVerdict {
	switch verdict {
	case refGranted:
		return refDenied
	case refDenied:
		return refGranted
	default:
		return refFailed
	}
}
