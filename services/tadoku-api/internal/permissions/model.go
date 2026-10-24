package permissions

import (
	"errors"
	"fmt"
)

// Model is a validated set of namespaces. Build it with NewModel in a composition root.
type Model struct {
	namespaces map[ObjectType]Namespace
}

// NewModel fails when two namespaces share an object type or any namespace refers to a relation, permit or
// object type the model does not define, so a broken model stops startup.
func NewModel(namespaces ...Namespace) (Model, error) {
	model := Model{namespaces: make(map[ObjectType]Namespace, len(namespaces))}
	var errs []error
	for _, namespace := range namespaces {
		if _, duplicate := model.namespaces[namespace.objectType]; duplicate {
			errs = append(errs, fmt.Errorf("object type %q is defined twice", namespace.objectType))
		}
		model.namespaces[namespace.objectType] = namespace
	}
	if err := errors.Join(append(errs, model.validate())...); err != nil {
		return Model{}, fmt.Errorf("permission model: %w", err)
	}
	return model, nil
}

func (m Model) namespace(objectType ObjectType) (Namespace, bool) {
	namespace, ok := m.namespaces[objectType]
	return namespace, ok
}
