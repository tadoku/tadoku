package permissions

import (
	"errors"
	"fmt"
	"slices"
)

func (m Model) validate() error {
	var errs []error
	for objectType, definition := range m {
		if objectType == userSubjectType {
			errs = append(errs, fmt.Errorf("object type %q is reserved for user subjects", objectType))
		}

		for relation, subjects := range definition.Relations {
			if _, clash := definition.Permits[relation]; clash {
				errs = append(errs, fmt.Errorf("%s#%s is both a relation and a permit", objectType, relation))
			}
			for _, subject := range subjects {
				if err := m.validateSubjectType(subject); err != nil {
					errs = append(errs, fmt.Errorf("%s#%s: %w", objectType, relation, err))
				}
			}
		}

		for permit, rule := range definition.Permits {
			if err := m.validateRule(objectType, rule); err != nil {
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

	definition, ok := m[subject.objectType]
	if !ok {
		return fmt.Errorf("unknown object type %q", subject.objectType)
	}
	_, relation := definition.Relations[subject.relation]
	_, permit := definition.Permits[subject.relation]
	if subject.relation != "" && !relation && !permit {
		return fmt.Errorf("unknown subject set %s#%s", subject.objectType, subject.relation)
	}
	return nil
}

func (m Model) validateRule(objectType ObjectType, rule Rule) error {
	definition := m[objectType]

	switch rule := rule.(type) {
	case relatedRule:
		subjects, ok := definition.Relations[rule.relation]
		if !ok {
			return fmt.Errorf("unknown relation %q", rule.relation)
		}
		if slices.ContainsFunc(subjects, SubjectType.isObject) {
			return fmt.Errorf("relation %q holds objects; use Via", rule.relation)
		}
	case allowedRule:
		if _, ok := definition.Permits[rule.permit]; !ok {
			return fmt.Errorf("unknown permit %q", rule.permit)
		}
	case appRule:
		if _, ok := m[App].Permits[rule.permit]; !ok {
			return fmt.Errorf("unknown app permit %q", rule.permit)
		}
	case viaRule:
		subjects, ok := definition.Relations[rule.relation]
		if !ok {
			return fmt.Errorf("unknown relation %q", rule.relation)
		}
		for _, subject := range subjects {
			if !subject.isObject() {
				return fmt.Errorf("relation %q holds users; use Related", rule.relation)
			}
			if _, ok := m[subject.objectType].Permits[rule.permit]; !ok {
				return fmt.Errorf("%s has no permit %q", subject.objectType, rule.permit)
			}
		}
	case anyOfRule:
		return m.validateRules(objectType, rule.rules)
	case allOfRule:
		return m.validateRules(objectType, rule.rules)
	case exceptRule:
		return errors.Join(m.validateRule(objectType, rule.base), m.validateRule(objectType, rule.subtract))
	default:
		return fmt.Errorf("unknown rule %T", rule)
	}
	return nil
}

func (m Model) validateRules(objectType ObjectType, rules []Rule) error {
	if len(rules) < 2 {
		return errors.New("combine at least two rules")
	}

	errs := make([]error, 0, len(rules))
	for _, rule := range rules {
		errs = append(errs, m.validateRule(objectType, rule))
	}
	return errors.Join(errs...)
}

func (m Model) validateRelationship(relationship Relationship) error {
	if relationship.object.id == "" {
		return errors.New("relationship object needs an ID")
	}

	subjects, ok := m[relationship.object.objectType].Relations[relationship.relation]
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
