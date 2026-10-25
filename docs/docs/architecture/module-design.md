---
title: Module design
description: How to prefer deep modules, hide implementation decisions, preserve meaningful boundaries, review the knowledge an interface requires of its callers and split a file over the size limit.
---

# Module design

Read this when you design or review a module boundary, public interface or
shared abstraction in backend or frontend code.

Prefer **deep modules**: a small, coherent interface that hides meaningful
implementation complexity. A module can be a package, service, component or
hook. Its interface includes not only signatures and props, but also the
ordering, error handling and assumptions a caller must know to use it correctly.
Judge depth by the burden on callers, not lines of code, file count or call-stack
depth. This follows John Ousterhout's
[A Philosophy of Software Design](https://web.stanford.edu/~ouster/cgi-bin/aposd.php)
and his discussion of
[information hiding](https://web.stanford.edu/~ouster/CS349W/lectures/abstraction.html).

## Put knowledge with its owner

- Expose the operation the caller needs, with domain inputs and results. Keep
  the owner's validation, normalization and internal sequencing behind that
  operation. Avoid requiring callers to run a separate validation method before
  invoking a write that accepts unchecked state.
- Keep decisions likely to change together inside one owner: storage mapping,
  cache keys and invalidation, provider adaptation or an interaction's mechanics.
  A change to an internal representation should not require coordinated edits
  across its consumers.
- Supply sensible defaults for ordinary use. Expose an option when callers
  have a real choice to make, not to move an unresolved implementation decision
  into every call site. Represent mutually exclusive choices with explicit
  variants and their required data.
- Extract a helper or component when its name and contract let a reader skip
  its implementation. Keep related steps together when splitting them would
  require following several calls to understand the same responsibility.
- Keep shared APIs to current needs. A general operation can simplify several
  existing callers; speculative extension points and configuration add to the
  interface before they hide any useful work.

## Preserve meaningful boundaries

A short function can enforce an important policy or translate between two
contracts. Retain it when it isolates authorization, a provider, a transaction
scope or a stable product convention. A pass-through is a review signal, not
an automatic reason to remove a layer. Do not add forwarding layers solely
for symmetry, hypothetical reuse or mocking.

Depth does not justify a module with unrelated responsibilities. Keep a
cohesive owner, use private helpers where they improve understanding, and
preserve the existing dependency and security rules.

| Area | Hide inside the owner | Keep explicit at the caller |
| --- | --- | --- |
| Backend feature | Feature validation, IDs and stored-state construction, and related row updates | Application actor authorization, composition of independent features and cross-feature locks and transactions |
| Repository or provider adapter | Generated row types, wire formats and provider-specific error translation | Domain inputs, results and meaningful failures |
| Frontend data hook | Requests, response parsing, query keys and invalidation inherent to its mutation | Screen navigation, messages and other product effects |
| UI control | Interaction mechanics, accessible structure and visual defaults | Application data, permissions and callbacks |

Feature services still own their repository call sequences; repository
methods retain the [single-statement rule](../tadoku-api/conventions.md#repositories-and-stores).
Keep [actor authorization](./authorization.md) in the application and preserve
[import boundaries](../tadoku-api/import-boundaries.md). Screens continue to own
product effects under [Paper composition](../frontend/paper-composition.md);
they need not know the data hook's cache representation to request a mutation.

A concern that many features use divides into core logic, state, vocabulary
and enforcement, each with one owner; see
[Cross-cutting concerns](./cross-cutting-concerns.md).

Document non-obvious caller obligations, such as an enclosing transaction or
concurrency constraint, at the declaration under the existing
[Go comment rules](../tadoku-api/conventions.md#readability). Make the contract
sufficient to use the module without reading its implementation.

## Review a boundary through its callers

Trace a real caller through the operation before deciding to add, split or
combine a module. Ask:

- What responsibility does the interface complete, and what knowledge does it
  let the caller forget?
- Must callers know storage fields, cache keys, provider details or a required
  sequence that belongs to this owner?
- Do several callers repeat that knowledge? Would changing one internal
  decision force all of them to change?
- Does each option represent a caller need? Could the ordinary path use a
  default or a more precise input type?
- If this boundary disappeared, what policy or translation would be lost?

A useful improvement removes an internal detail or required step from callers
while preserving observable behavior and ownership. Verify affected workflows
using [Verifying changes](../develop/verifying-changes.md). Do not use quotas
for function size or exported methods as a substitute for this review. The file
size limit below prompts a split; it does not choose the boundary.

## Splitting a long file

`./tools/ci/check_file_sizes.sh` fails when a handwritten text file is over
20,000 bytes, about 5,000 tokens. Agents and reviewers read a file whole, so a
long file costs them context before they reach the part they need.

- Split along responsibilities the file already has: a component with its
  documentation and fixtures, a domain's schemas and hooks, the tests for one
  behavior, or a self-contained docs topic.
- Keep the interface. A package keeps its public exports, a Go package keeps
  its name and visibility, and a docs page keeps its URL. Inside an
  application, import from the new module instead of adding a re-export file.
- Keep each operation readable in one place. Do not split by line count or
  alphabetically, and do not add pass-through files.
- In Go, move declarations or tests into another file in the same package. Put
  test fixtures shared by several test files in a `_test.go` helper file.

Generated files and lockfiles are exempt when `.gitattributes` marks them
`linguist-generated=true`. Every other text file is held to the limit.
