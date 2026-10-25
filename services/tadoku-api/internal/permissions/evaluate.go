package permissions

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const maxDepth = 8

var errMaxDepth = errors.New("permission check exceeded the maximum depth")

type question struct {
	check Check
	user  uuid.UUID
}

// evaluate answers the questions together, one rule level at a time: each round walks every undecided question over
// the relations read so far and reads all the relations they still need with one Subjects call. A storage error, or a
// question an error leaves undecided, such as a walk past maxDepth, fails the whole call. An error in a rule that
// cannot change the answer does not.
func (m Model) evaluate(ctx context.Context, tuples Tuples, questions []question) ([]bool, error) {
	e := evaluation{model: m, read: map[SubjectsKey]subjects{}, answers: map[answerKey]verdict{}}
	answers := make([]bool, len(questions))
	open := make([]int, len(questions))
	for i := range open {
		open[i] = i
	}

	for {
		e.wanted = e.wanted[:0]
		undecided := open[:0]
		for _, i := range open {
			question := questions[i]
			v := e.permit(question.check.object, question.check.permit, question.user, 0)
			switch v.state {
			case stateFailed:
				return nil, v.err
			case statePending:
				undecided = append(undecided, i)
			default:
				answers[i] = v.state == stateAllowed
			}
		}
		if len(undecided) == 0 {
			return answers, nil
		}

		if err := e.readWanted(ctx, tuples); err != nil {
			return nil, err
		}
		open = undecided
	}
}

type state uint8

const (
	stateDenied state = iota
	stateAllowed
	statePending
	stateFailed
)

// verdict is the answer to one question so far. A pending verdict waits for a relation that is not read yet; err is
// set only on a failed verdict.
type verdict struct {
	state state
	err   error
}

type answerKey struct {
	object Object
	name   string
	user   uuid.UUID
	depth  int
}

type subjects struct {
	direct  bool
	sets    []subjectSet
	targets []Object
}

// evaluation remembers, for one call, the relations read and every decided answer, so an identical sub-question is
// answered once. wanted holds the reads the undecided part of the current round needs.
type evaluation struct {
	model   Model
	read    map[SubjectsKey]subjects
	answers map[answerKey]verdict
	wanted  []SubjectsKey
}

// remember returns the decided answer for key, or decides it and drops the reads that deciding it registered.
func (e *evaluation) remember(key answerKey, decide func() verdict) verdict {
	if v, ok := e.answers[key]; ok {
		return v
	}

	mark := len(e.wanted)
	v := decide()
	if v.state != statePending {
		e.answers[key] = v
		e.wanted = e.wanted[:mark]
	}
	return v
}

func (e *evaluation) permit(object Object, permit string, user uuid.UUID, depth int) verdict {
	return e.remember(answerKey{object: object, name: permit, user: user, depth: depth}, func() verdict {
		if depth > maxDepth {
			return verdict{state: stateFailed, err: errMaxDepth}
		}

		namespace, _ := e.model.namespace(object.objectType)
		rule, ok := namespace.permits[permit]
		if !ok {
			return verdict{state: stateFailed, err: fmt.Errorf("unknown permit %s#%s", object.objectType, permit)}
		}
		return e.rule(object, rule, user, depth)
	})
}

func (e *evaluation) rule(object Object, rule Rule, user uuid.UUID, depth int) verdict {
	mark := len(e.wanted)
	v := e.decideRule(object, rule, user, depth)
	if v.state != statePending {
		e.wanted = e.wanted[:mark]
	}
	return v
}

func (e *evaluation) decideRule(object Object, rule Rule, user uuid.UUID, depth int) verdict {
	switch rule := rule.(type) {
	case relatedRule:
		return e.relation(object, rule.relation, user, depth+1)
	case allowedRule:
		return e.permit(object, rule.permit, user, depth+1)
	case permitOnRule:
		return e.permit(rule.object, rule.permit, user, depth+1)
	case viaRule:
		return e.via(object, rule, user, depth+1)
	case anyOfRule:
		return anyOf(len(rule.rules), func(i int) verdict { return e.rule(object, rule.rules[i], user, depth) })
	case allOfRule:
		return allOf(len(rule.rules), func(i int) verdict { return e.rule(object, rule.rules[i], user, depth) })
	case exceptRule:
		return e.except(object, rule, user, depth)
	default:
		return verdict{state: stateFailed, err: fmt.Errorf("unknown rule %T", rule)}
	}
}

func (e *evaluation) relation(object Object, relation string, user uuid.UUID, depth int) verdict {
	return e.remember(answerKey{object: object, name: relation, user: user, depth: depth}, func() verdict {
		if depth > maxDepth {
			return verdict{state: stateFailed, err: errMaxDepth}
		}

		stored, ok := e.lookup(object, relation, user)
		if !ok {
			return verdict{state: statePending}
		}
		if stored.direct {
			return verdict{state: stateAllowed}
		}
		return anyOf(len(stored.sets), func(i int) verdict { return e.subjectSet(stored.sets[i], user, depth+1) })
	})
}

func (e *evaluation) subjectSet(set subjectSet, user uuid.UUID, depth int) verdict {
	namespace, _ := e.model.namespace(set.object.objectType)
	if _, ok := namespace.permits[set.relation]; ok {
		return e.permit(set.object, set.relation, user, depth)
	}
	return e.relation(set.object, set.relation, user, depth)
}

func (e *evaluation) via(object Object, rule viaRule, user uuid.UUID, depth int) verdict {
	if depth > maxDepth {
		return verdict{state: stateFailed, err: errMaxDepth}
	}

	stored, ok := e.lookup(object, rule.relation, user)
	if !ok {
		return verdict{state: statePending}
	}
	return anyOf(len(stored.targets), func(i int) verdict {
		return e.permit(stored.targets[i], rule.permit, user, depth)
	})
}

func (e *evaluation) except(object Object, rule exceptRule, user uuid.UUID, depth int) verdict {
	base := e.rule(object, rule.base, user, depth)
	if base.state == stateDenied {
		return base
	}

	subtract := e.rule(object, rule.subtract, user, depth)
	switch {
	case subtract.state == stateAllowed:
		return verdict{state: stateDenied}
	case base.state == statePending || subtract.state == statePending:
		return verdict{state: statePending}
	case base.state == stateFailed:
		return base
	case subtract.state == stateFailed:
		return subtract
	default:
		return verdict{state: stateAllowed}
	}
}

// anyOf passes when any verdict passes. Otherwise a pending verdict keeps it pending, then a failed one fails it.
func anyOf(count int, verdictOf func(int) verdict) verdict {
	result := verdict{state: stateDenied}
	for i := range count {
		v := verdictOf(i)
		switch {
		case v.state == stateAllowed:
			return v
		case v.state == statePending:
			result = v
		case v.state == stateFailed && result.state == stateDenied:
			result = v
		}
	}
	return result
}

// allOf denies when any verdict denies. Otherwise a pending verdict keeps it pending, then a failed one fails it.
func allOf(count int, verdictOf func(int) verdict) verdict {
	result := verdict{state: stateAllowed}
	for i := range count {
		v := verdictOf(i)
		switch {
		case v.state == stateDenied:
			return v
		case v.state == statePending:
			result = v
		case v.state == stateFailed && result.state == stateAllowed:
			result = v
		}
	}
	return result
}

func (e *evaluation) lookup(object Object, relation string, user uuid.UUID) (subjects, bool) {
	key := SubjectsKey{
		ObjectType: string(object.objectType),
		ObjectID:   object.id,
		Relation:   relation,
		UserID:     user.String(),
	}
	stored, ok := e.read[key]
	if !ok {
		e.wanted = append(e.wanted, key)
	}
	return stored, ok
}

// readWanted reads every relation the round wants with one Subjects call. A round that wants nothing new while a
// question is undecided would never end, so it fails.
func (e *evaluation) readWanted(ctx context.Context, tuples Tuples) error {
	var keys []SubjectsKey
	byRelation := map[SubjectsKey][]SubjectsKey{}
	for _, key := range e.wanted {
		if _, done := e.read[key]; done {
			continue
		}
		e.read[key] = subjects{}
		keys = append(keys, key)
		relation := SubjectsKey{ObjectType: key.ObjectType, ObjectID: key.ObjectID, Relation: key.Relation}
		byRelation[relation] = append(byRelation[relation], key)
	}
	if len(keys) == 0 {
		return errors.New("permission check made no progress")
	}

	stored, err := tuples.Subjects(ctx, keys)
	if err != nil {
		return err
	}

	for _, tuple := range stored {
		relation := SubjectsKey{ObjectType: tuple.ObjectType, ObjectID: tuple.ObjectID, Relation: tuple.Relation}
		subject := Object{objectType: ObjectType(tuple.SubjectType), id: tuple.SubjectID}
		for _, key := range byRelation[relation] {
			found := e.read[key]
			switch {
			case tuple.SubjectType == UserSubjectType:
				found.direct = found.direct || tuple.SubjectID == key.UserID
			case tuple.SubjectRelation == "":
				found.targets = append(found.targets, subject)
			default:
				found.sets = append(found.sets, subjectSet{object: subject, relation: tuple.SubjectRelation})
			}
			e.read[key] = found
		}
	}
	return nil
}
