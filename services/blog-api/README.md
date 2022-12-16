## Blog

Currently just a stub service that returns a hardcoded JSON response.

## TODO docs

```sh
# Generate new api client
bazel run //services/blog-api:api_gen

# Generate new version of OpenAPI client
bazel run //services/blog-api/http/rest/openapi:codegen_gen
```