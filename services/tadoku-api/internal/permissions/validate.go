package permissions

import (
	"errors"
	"fmt"
	"slices"
)

func (m Model) validate() error {
	var errs []error
	for objectType, namespace := range m.namespaces {
		switch objectType {
		case "":
			errs = append(errs, errors.New("a namespace needs an object type"))
		case userSubjectType:
			errs = append(errs, fmt.Errorf("object type %q is reserved for user subjects", objectType))
		}

		for relation, subjects := range namespace.relations {
			if _, clash := namespace.permits[relation]; clash {
				errs = append(errs, fmt.Errorf("%s#%s is both a relation and a permit", objectType, relation))
			}
			for _, subject := range subjects {
				if err := m.validateSubjectType(subject); err != nil {
					errs = append(errs, fmt.Errorf("%s#%s: %w", objectType, relation, err))
				}
			}
		}

		for permit, rule := range namespace.permits {
			if err := m.validateRule(namespace, rule); err != nil {
				errs = append(errs, fmt.Errorf("%s#%s: %w", objectType, permit, err))
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
	_, relation := namespace.relations[subject.relation]
	_, permit := namespace.permits[subject.relation]
	if subject.relation != "" && !relation && !permit {
		return fmt.Errorf("unknown subject set %s#%s", subject.objectType, subject.relation)
	}
	return nil
}

func (m Model) validateRule(namespace Namespace, rule Rule) error {
	switch rule := rule.(type) {
	case relatedRule:
		subjects, ok := namespace.relations[rule.relation]
		if !ok {
			return fmt.Errorf("unknown relation %q", rule.relation)
		}
		if slices.ContainsFunc(subjects, SubjectType.isObject) {
			return fmt.Errorf("relation %q holds objects; use Via", rule.relation)
		}
	case allowedRule:
		if _, ok := namespace.permits[rule.permit]; !ok {
			return fmt.Errorf("unknown permit %q", rule.permit)
		}
	case permitOnRule:
		if rule.object.id == "" {
			return fmt.Errorf("PermitOn %s needs an object ID", rule.object.objectType)
		}
		target, ok := m.namespace(rule.object.objectType)
		if !ok {
			return fmt.Errorf("unknown object type %q", rule.object.objectType)
		}
		if _, ok := target.permits[rule.permit]; !ok {
			return fmt.Errorf("%s has no permit %q", rule.object.objectType, rule.permit)
		}
	case viaRule:
		subjects, ok := namespace.relations[rule.relation]
		if !ok {
			return fmt.Errorf("unknown relation %q", rule.relation)
		}
		for _, subject := range subjects {
			if !subject.isObject() {
				return fmt.Errorf("relation %q holds users; use Related", rule.relation)
			}
			target, _ := m.namespace(subject.objectType)
			if _, ok := target.permits[rule.permit]; !ok {
				return fmt.Errorf("%s has no permit %q", subject.objectType, rule.permit)
			}
		}
	case anyOfRule:
		return m.validateRules(namespace, rule.rules)
	case allOfRule:
		return m.validateRules(namespace, rule.rules)
	case exceptRule:
		return errors.Join(m.validateRule(namespace, rule.base), m.validateRule(namespace, rule.subtract))
	default:
		return fmt.Errorf("unknown rule %T", rule)
	}
	return nil
}

func (m Model) validateRules(namespace Namespace, rules []Rule) error {
	if len(rules) < 2 {
		return errors.New("combine at least two rules")
	}

	errs := make([]error, 0, len(rules))
	for _, rule := range rules {
		errs = append(errs, m.validateRule(namespace, rule))
	}
	return errors.Join(errs...)
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
	case subjectSet:
		if subject.object.id == "" {
			return errors.New("relationship subject needs an object ID")
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
