package permissions

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const maxDepth = 8

var errMaxDepth = errors.New("permission check exceeded the maximum depth")

// check fails closed: a storage error, including a cancelled context, an unknown permit or a walk deeper than
// maxDepth returns false with an error.
func (m Model) check(ctx context.Context, tuples Tuples, check Check, user uuid.UUID) (bool, error) {
	evaluation := evaluation{model: m, tuples: tuples, user: user}
	return evaluation.permit(ctx, check.object, check.permit, 0)
}

type evaluation struct {
	model  Model
	tuples Tuples
	user   uuid.UUID
}

func (e evaluation) permit(ctx context.Context, obj object, permit string, depth int) (bool, error) {
	if depth > maxDepth {
		return false, errMaxDepth
	}

	namespace, _ := e.model.namespace(obj.objectType)
	rule, ok := namespace.permits[permit]
	if !ok {
		return false, fmt.Errorf("unknown permit %s#%s", obj.objectType, permit)
	}
	return e.rule(ctx, obj, rule, depth)
}

func (e evaluation) rule(ctx context.Context, obj object, rule Rule, depth int) (bool, error) {
	switch rule := rule.(type) {
	case relatedRule:
		return e.relation(ctx, obj, rule.relation, depth+1)
	case allowedRule:
		return e.permit(ctx, obj, rule.permit, depth+1)
	case permitOnRule:
		return e.permit(ctx, rule.object, rule.permit, depth+1)
	case viaRule:
		return e.via(ctx, obj, rule, depth+1)
	case anyOfRule:
		return e.anyOf(ctx, obj, rule.rules, depth)
	case allOfRule:
		return e.allOf(ctx, obj, rule.rules, depth)
	case exceptRule:
		return e.except(ctx, obj, rule, depth)
	default:
		return false, fmt.Errorf("unknown rule %T", rule)
	}
}

func (e evaluation) relation(ctx context.Context, obj object, relation string, depth int) (bool, error) {
	if depth > maxDepth {
		return false, errMaxDepth
	}

	direct, sets, err := storedSubjects(ctx, e.tuples, obj, relation, e.user)
	if err != nil {
		return false, err
	}
	if direct {
		return true, nil
	}

	for _, set := range sets {
		passed, err := e.subjectSet(ctx, set, depth+1)
		if err != nil {
			return false, err
		}
		if passed {
			return true, nil
		}
	}
	return false, nil
}

func (e evaluation) subjectSet(ctx context.Context, set subjectSet, depth int) (bool, error) {
	namespace, _ := e.model.namespace(set.object.objectType)
	if _, ok := namespace.permits[set.relation]; ok {
		return e.permit(ctx, set.object, set.relation, depth)
	}
	return e.relation(ctx, set.object, set.relation, depth)
}

func (e evaluation) via(ctx context.Context, obj object, rule viaRule, depth int) (bool, error) {
	if depth > maxDepth {
		return false, errMaxDepth
	}

	_, targets, err := storedSubjects(ctx, e.tuples, obj, rule.relation, e.user)
	if err != nil {
		return false, err
	}

	for _, target := range targets {
		passed, err := e.permit(ctx, target.object, rule.permit, depth)
		if err != nil {
			return false, err
		}
		if passed {
			return true, nil
		}
	}
	return false, nil
}

func (e evaluation) anyOf(ctx context.Context, obj object, rules []Rule, depth int) (bool, error) {
	for _, rule := range rules {
		passed, err := e.rule(ctx, obj, rule, depth)
		if err != nil {
			return false, err
		}
		if passed {
			return true, nil
		}
	}
	return false, nil
}

func (e evaluation) allOf(ctx context.Context, obj object, rules []Rule, depth int) (bool, error) {
	for _, rule := range rules {
		passed, err := e.rule(ctx, obj, rule, depth)
		if err != nil || !passed {
			return false, err
		}
	}
	return true, nil
}

func (e evaluation) except(ctx context.Context, obj object, rule exceptRule, depth int) (bool, error) {
	passed, err := e.rule(ctx, obj, rule.base, depth)
	if err != nil || !passed {
		return false, err
	}

	excluded, err := e.rule(ctx, obj, rule.subtract, depth)
	if err != nil {
		return false, err
	}
	return !excluded, nil
}
