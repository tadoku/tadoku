package permissions

import (
	"errors"
	"fmt"
	"slices"
)

func (n Namespace) validate() error {
	var errs []error
	switch n.objectType {
	case "":
		errs = append(errs, errors.New("a namespace needs an object type"))
	case UserSubjectType:
		errs = append(errs, fmt.Errorf("object type %q is reserved for user subjects", n.objectType))
	}

	for relation, subjects := range n.relations {
		if _, clash := n.permits[relation]; clash {
			errs = append(errs, fmt.Errorf("%s is both a relation and a permit", relation))
		}
		for _, subject := range subjects {
			local := subject.objectType == n.objectType && subject.relation != ""
			if local && !n.defines(subject.relation) {
				errs = append(errs, fmt.Errorf("%s: unknown subject set %s", relation, subject.relation))
			}
		}
	}

	for permit, rule := range n.permits {
		leaves, err := leafRules(rule)
		errs = append(errs, err)
		for _, leaf := range leaves {
			if err := n.validateRule(leaf); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", permit, err))
			}
		}
	}
	return errors.Join(errs...)
}

func (n Namespace) defines(name string) bool {
	_, relation := n.relations[name]
	_, permit := n.permits[name]
	return relation || permit
}

// leafRules flattens the combinators in rule and checks that each AnyOf and AllOf combines at least two rules.
func leafRules(rule Rule) ([]Rule, error) {
	var children []Rule
	switch rule := rule.(type) {
	case anyOfRule:
		children = rule.rules
	case allOfRule:
		children = rule.rules
	case exceptRule:
		children = []Rule{rule.base, rule.subtract}
	default:
		return []Rule{rule}, nil
	}

	var leaves []Rule
	var errs []error
	if len(children) < 2 {
		errs = append(errs, errors.New("combine at least two rules"))
	}
	for _, child := range children {
		childLeaves, err := leafRules(child)
		leaves = append(leaves, childLeaves...)
		errs = append(errs, err)
	}
	return leaves, errors.Join(errs...)
}

func (n Namespace) validateRule(rule Rule) error {
	switch rule := rule.(type) {
	case relatedRule:
		subjects, ok := n.relations[rule.relation]
		if !ok {
			return fmt.Errorf("unknown relation %q", rule.relation)
		}
		if slices.ContainsFunc(subjects, SubjectType.isObject) {
			return fmt.Errorf("relation %q holds objects; use Via", rule.relation)
		}
	case allowedRule:
		if _, ok := n.permits[rule.permit]; !ok {
			return fmt.Errorf("unknown permit %q", rule.permit)
		}
	case viaRule:
		subjects, ok := n.relations[rule.relation]
		if !ok {
			return fmt.Errorf("unknown relation %q", rule.relation)
		}
		if slices.ContainsFunc(subjects, func(subject SubjectType) bool { return !subject.isObject() }) {
			return fmt.Errorf("relation %q holds users; use Related", rule.relation)
		}
	case permitOnRule:
		if rule.object.id == "" {
			return fmt.Errorf("PermitOn %s needs an object ID", rule.object.objectType)
		}
	default:
		return fmt.Errorf("unknown rule %T", rule)
	}
	return nil
}

func (m Model) validateReferences(n Namespace) error {
	var errs []error
	for relation, subjects := range n.relations {
		for _, subject := range subjects {
			if err := m.validateSubjectType(subject); err != nil {
				errs = append(errs, fmt.Errorf("%s#%s: %w", n.objectType, relation, err))
			}
		}
	}

	for permit, rule := range n.permits {
		leaves, _ := leafRules(rule)
		for _, leaf := range leaves {
			if err := m.validateRuleReferences(n, leaf); err != nil {
				errs = append(errs, fmt.Errorf("%s#%s: %w", n.objectType, permit, err))
			}
		}
	}
	return errors.Join(errs...)
}

func (m Model) validateSubjectType(subject SubjectType) error {
	if subject == Users {
		return nil
	}

	namespace, ok := m.namespace(subject.objectType)
	if !ok {
		return fmt.Errorf("unknown object type %q", subject.objectType)
	}
	if subject.relation != "" && !namespace.defines(subject.relation) {
		return fmt.Errorf("unknown subject set %s#%s", subject.objectType, subject.relation)
	}
	return nil
}

func (m Model) validateRuleReferences(n Namespace, rule Rule) error {
	switch rule := rule.(type) {
	case permitOnRule:
		target, ok := m.namespace(rule.object.objectType)
		if !ok {
			return fmt.Errorf("unknown object type %q", rule.object.objectType)
		}
		if _, ok := target.permits[rule.permit]; !ok {
			return fmt.Errorf("%s has no permit %q", rule.object.objectType, rule.permit)
		}
	case viaRule:
		for _, subject := range n.relations[rule.relation] {
			target, _ := m.namespace(subject.objectType)
			if _, ok := target.permits[rule.permit]; !ok {
				return fmt.Errorf("%s has no permit %q", subject.objectType, rule.permit)
			}
		}
	}
	return nil
}

func (m Model) validateObject(object Object) error {
	if _, ok := m.namespace(object.objectType); !ok {
		return fmt.Errorf("unknown object type %q", object.objectType)
	}
	if object.id == "" {
		return errors.New("object needs an ID")
	}
	return nil
}

func (m Model) validateRelationship(relationship Relationship) error {
	if relationship.object.id == "" {
		return errors.New("relationship object needs an ID")
	}

	namespace, _ := m.namespace(relationship.object.objectType)
	subjects, ok := namespace.relations[relationship.relation]
	if !ok {
		return fmt.Errorf("unknown relation %s#%s", relationship.object.objectType, relationship.relation)
	}

	var subjectType SubjectType
	switch subject := relationship.subject.(type) {
	case userSubject:
		subjectType = Users
	case Object:
		if subject.objectType == "" || subject.id == "" {
			return errors.New("linked object needs an object type and an ID")
		}
		subjectType = ObjectOf(subject.objectType)
	case subjectSet:
		if subject.object.objectType == "" || subject.object.id == "" || subject.relation == "" {
			return errors.New("subject set needs an object type, an ID and a relation")
		}
		subjectType = SetOf(subject.object.objectType, subject.relation)
	default:
		return errors.New("relationship needs a subject")
	}
	if !slices.Contains(subjects, subjectType) {
		return fmt.Errorf(
			"%s#%s does not accept %+v",
			relationship.object.objectType,
			relationship.relation,
			relationship.subject,
		)
	}
	return nil
}
