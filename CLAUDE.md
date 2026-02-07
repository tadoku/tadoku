# Development Workflow

## Frontend

**Always use `pnpm`, not `npm`.**

**Use the `ui` package design system** - never write custom button/form styles. Use:
- Buttons: `className="btn"` with variants `primary`, `secondary`, `danger`, `ghost`
- Forms: Import `Input`, `Select`, `Checkbox`, etc. from `ui/components/Form`
- Components: Import from `ui` package (Modal, Flash, Navbar, etc.)

**Always use `react-hook-form` for form handling** — never use plain `useState` for form fields. Use:
- `useForm()` + `<FormProvider>` to set up form context
- `<Input>`, `<Select>`, `<TextArea>` from the `ui` package (they use `useFormContext()` internally)
- `useController()` for custom/non-standard form components (e.g. CodeEditor)
- `methods.handleSubmit()` for form submission with built-in validation
- `methods.watch()` for reactive field values (e.g. live previews)
- `methods.reset()` to populate forms with existing data

```sh
# 1. Make changes

# 2. Typecheck (fast)
cd frontend && pnpm --filter webv2 exec tsc --noEmit

# 3. Lint before committing
cd frontend && pnpm --filter webv2 lint

# 4. Before creating PR
cd frontend && pnpm build
```

## Backend

**Always use `bazel`, not `go`.**

**Always write tests for new backend functionality** — new domain services, repository methods, and HTTP handlers should have corresponding test coverage.

**Always use `testify` for test assertions** — use `assert` for checks and `require` for fatal preconditions (`github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`). Never use raw `if err != nil { t.Fatal(...) }` patterns.

**SQL style: always use lowercase keywords** (select, create table, not SELECT, CREATE TABLE)

```sh
# 1. Make changes

# 2. Compile (fast)
bazel build //services/...

# 3. Run tests
bazel test //services/... # everything
bazel test //services/immersion-api/domain/command:command_test # one test file
bazel test //services/immersion-api/domain/command:command_test --test_filter=TestValidateAndNormalizeTags # specific function

# 4. Format before committing
gofmt -w services/

# 5. Regenerate BUILD.bazel files (after adding/removing Go files or changing deps/imports)
bazel run //:gazelle

# 6. Regenerate sqlc code (after modifying SQL queries)
cd services/immersion-api/storage/postgres && go generate
cd services/content-api/storage/postgres && go generate

# 7. Regenerate OpenAPI code (after modifying OpenAPI specs)
cd services/immersion-api/http/rest/openapi && go generate
cd services/content-api/http/rest/openapi && go generate

# 8. Before creating PR
bazel build //services/... && bazel test //services/...
```

## Commit Guidelines

**Commit in atomic diffs** — each commit should represent one logical change. Don't bundle unrelated changes into a single commit.

**For larger refactors spanning many files**, commit in chunks that make sense — e.g. one commit per page, per service, per domain area, etc.

## Bug Reports

When a bug is reported, follow this process:

1. **Write a failing test first** — Don't start by trying to fix the bug. Instead, write a test that reproduces the bug and confirms it fails.
2. **Use subagents to fix** — Have subagents attempt to fix the bug and prove the fix by making the test pass.
