---
title: Permission model
description: How a Tadoku API feature defines its permission namespace and handles, how relationships are stored, and the building blocks for checks, writes, cleanup and reads.
---

# Permission model

Read this when you define a feature's permissions, write or read its
relationships, or check a permit.

The model follows
[Zanzibar](https://research.google/pubs/zanzibar-googles-consistent-global-authorization-system/).
Relation tuples are rows in the tenant-owned table `permission_relationships`;
row-level security scopes them to the transaction's tenant, and deleting a
tenant deletes them. A subject is a user (`user` with the lowercase account UUID
from the JWT `sub`, never an email), a linked object, or a subject set, which
names a relation or a permit of an object. The `app` namespace and its roles are
in [Authentication and authorization](./authorization.md#permission-model).

## Namespaces and handles

The core logic in `services/tadoku-api/internal/permissions/` knows no object
type. Each feature defines the namespace of the objects it owns with
`permissions.MustDefine` in `<feature>_permissions.go`, with permits built from
these rules:

| Rule | Zanzibar equivalent |
| --- | --- |
| `Related(relation)` | `_this`, including subject sets |
| `Allowed(permit)` | `computed_userset` |
| `Via(relation, permit)` | `tuple_to_userset` |
| `AnyOf`, `AllOf`, `Except` | Union, intersection, exclusion |
| `PermitOn(objectType, objectID, permit)` | None: the permit on one fixed object, which may belong to another feature's namespace |

Features cannot import each other, so a reference to another namespace is a
name. A feature turns each name it uses into a handle once, in a package-level
variable, and builds every check and relationship from handles. No other code
can build them, and Bazel visibility keeps other features from importing a
namespace. An example, not existing code, where a log's owner or a moderator of
its contest may edit it:

```go
var Permissions = permissions.MustDefine(
	"log",
	permissions.Relations{
		"owner":   {permissions.Users},
		"contest": {permissions.ObjectOf("contest")},
	},
	permissions.Permits{
		"edit": permissions.AnyOf(permissions.Related("owner"), permissions.Via("contest", "moderate")),
	},
)

var (
	edit    = Permissions.MustPermit("edit")
	owner   = Permissions.MustUserRelation("owner")
	contest = Permissions.MustObjectRelation("contest")
)

func EditPermission(logID uuid.UUID) permissions.Check {
	return edit.On(logID.String())
}

func OwnerRelationship(logID, userID uuid.UUID) permissions.Relationship {
	return owner.On(logID.String(), userID)
}

func ContestRelationship(logID, contestID uuid.UUID) permissions.Relationship {
	return contest.On(logID.String(), contestID.String())
}
```

| Handle | Names | `On` builds |
| --- | --- | --- |
| `MustPermit(name)` | A permit | A `Check` on an object |
| `MustUserRelation(name)` | A relation that accepts `Users` | The user in the relation |
| `MustObjectRelation(name)` | A relation that accepts one `ObjectOf` type; `MustObjectRelationOf(name, permissions.ObjectOf(t))` picks one of several | A link to an object, which `Via` follows |
| `MustSetRelation(name)` | A relation that accepts one `SetOf` subject set; `MustSetRelationOf(name, permissions.SetOf(t, r))` picks one of several | The users in a relation or permit of an object |

Exported helpers take typed IDs, so call sites never spell a namespace, permit
or relation: `<Permit>Permission` returns a `permissions.Check` and
`<Relation>Relationship` returns a `permissions.Relationship`. The `authz`
helpers stay unexported behind `authz.Roles`.

A rule or handle that names a missing permit or relation, or a relation with
the wrong kind of subject, panics when the package loads. Every composition
root builds its model with `permissionmodel.New` from the namespace list in
`services/tadoku-api/app/permissionmodel/permissionmodel.go`, which rejects a
wrong reference across namespaces: its test fails CI and `tadoku-api` does not
start. To add a namespace, add it to that list. Every check, write and read
rejects a hand-built value, such as a zero `permissions.Check`.

## Building blocks

Every building block runs in the context tenant, joins the context's
transaction (`postgres.RunInTransaction`) and fails closed.

| Building block | Use | SQL statements |
| --- | --- | --- |
| `Checker.Require`, `Allowed` and `AllowedFor` | Check one permit for the actor or for a given user | At most one for each rule level reached |
| `Checker.AllowedEach` | Check many permits for the actor, such as whether each row on a page is editable | At most one for each rule level reached, whatever the number of checks |
| `Checker.AllowedForEach` | Check one permit for many users, such as role badges on a user list | At most one for each rule level reached, whatever the number of users |
| `Relationships.Grant` and `Revoke` | Store or remove any number of relationships | 1 |
| `Relationships.RevokeObject` and `RevokeUser` | Remove every relationship of a deleted object or user | 1 |
| `Relationships.Read` | Read stored relationships on an object or for a user | 1 |

[Checks in application operations](./authorization.md#checks-in-application-operations)
describes the checker's results and errors. `AllowedForEach`, like
`AllowedFor`, reads no request identity and never authorizes the actor.

A call evaluates all its questions together, one rule level at a time, reads
every relation a level needs with one query, and answers an identical
sub-question once. An error fails the whole call. A rule that cannot be
decided, such as a walk deeper than eight levels, fails a check only when the
answer depends on it.

`Grant` and `Revoke` validate every relationship before any SQL. Granting a
stored relationship or revoking a missing one changes nothing.

`RevokeObject` removes every relationship on an object and every relationship
that links to it or names it in a subject set. `RevokeUser` removes every
relationship whose subject is the user, but not subject sets that contain the
user. Call `RevokeObject` from `app/` in the transaction that deletes the
object; `Namespace.Object(objectID)` names the object.

`Read` returns stored relationships, sorted by object, relation and subject,
without evaluating rules; use a check for anything that depends on rules. Its
filter starts from an object or a user:

| Filter | Selects |
| --- | --- |
| `permissions.OnObject(object)` | The relationships on the object |
| `permissions.ForUser(userID)` | The relationships whose subject is the user, not those through a subject set |
| `.Relation(handle)` | Only that relation |
| `.In(namespace)` | For `ForUser`, only objects of that namespace |

`Read` rejects an object or namespace the model lacks, `In` on an object
filter, and a relation handle of another namespace.

The core logic reaches storage only through its `permissions.Tuples` port,
which the `features/authz` repository implements
(`services/tadoku-api/features/authz/authz_repository.go`).
