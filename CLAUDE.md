# Development Workflow

## Frontend

**Always use `pnpm`, not `npm`.**

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
