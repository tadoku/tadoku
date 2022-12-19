## Content-API

## TODO docs

```sh
# Generate new api client
bazel run //services/content-api:api_gen

# Generate new version of OpenAPI client
bazel run //services/content-api/http/rest/openapi:codegen_gen
```