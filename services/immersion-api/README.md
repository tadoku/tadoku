## Immersion-API

## TODO docs

```sh
# Generate new api client
bazel run //services/immersion-api:api_gen

# Generate new version of OpenAPI client
bazel run //services/immersion-api/http/rest/openapi:codegen_gen
```