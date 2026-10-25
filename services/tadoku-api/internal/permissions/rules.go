package permissions

type Rule interface {
	rule()
}

type relatedRule struct{ relation string }
type allowedRule struct{ permit string }
type viaRule struct{ relation, permit string }
type permitOnRule struct {
	object Object
	permit string
}
type anyOfRule struct{ rules []Rule }
type allOfRule struct{ rules []Rule }
type exceptRule struct{ base, subtract Rule }

func (relatedRule) rule()  {}
func (allowedRule) rule()  {}
func (viaRule) rule()      {}
func (permitOnRule) rule() {}
func (anyOfRule) rule()    {}
func (allOfRule) rule()    {}
func (exceptRule) rule()   {}

// Related passes for users stored in the relation, directly or through a subject set.
func Related(relation string) Rule { return relatedRule{relation: relation} }

// Allowed passes when another permit on the same object passes.
func Allowed(permit string) Rule { return allowedRule{permit: permit} }

// Via passes when the permit passes on any object stored in the relation.
func Via(relation, permit string) Rule { return viaRule{relation: relation, permit: permit} }

// PermitOn passes when the permit passes on one fixed object, which may belong to another feature's namespace.
func PermitOn(objectType ObjectType, objectID, permit string) Rule {
	return permitOnRule{object: Object{objectType: objectType, id: objectID}, permit: permit}
}

func AnyOf(rules ...Rule) Rule { return anyOfRule{rules: rules} }

func AllOf(rules ...Rule) Rule { return allOfRule{rules: rules} }

func Except(base, subtract Rule) Rule { return exceptRule{base: base, subtract: subtract} }

// SubjectType is one kind of subject a relation accepts: Users, SetOf or ObjectOf.
type SubjectType struct {
	objectType ObjectType
	relation   string
}

var Users = SubjectType{}

func SetOf(objectType ObjectType, relation string) SubjectType {
	return SubjectType{objectType: objectType, relation: relation}
}

// ObjectOf accepts linked objects of the type, which Via follows.
func ObjectOf(objectType ObjectType) SubjectType {
	return SubjectType{objectType: objectType}
}

func (subject SubjectType) isObject() bool {
	return subject.objectType != "" && subject.relation == ""
}

func (subject SubjectType) isSet() bool {
	return subject.objectType != "" && subject.relation != ""
}

type Relations map[string][]SubjectType

type Permits map[string]Rule
