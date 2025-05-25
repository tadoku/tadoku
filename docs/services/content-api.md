# Content API

This is the blog/content API of the Tadoku website. It contains the functionality to:

- Fetch/Create/Update/Delete blog posts
- Fetch/Create/Update/Delete web pages

All content uses markdown.

## Architecture

- This service is written in Golang.
- OpenAPI REST API over HTTP
  - Spec can be found at `services/content-api/http/rest/openapi/api.yaml`
- Uses CQRS (Command Query Responsibility Segregation) to split data read/writes
- Data is stored in a PostgreSQL database
- sqlc is used to generate Golang glue code for raw sql queries.
  - New queries should be written in `services/content-api/storage/postgres/queries`
  - Run the following command to generate code for the query: `cd services/content-api/storage/postgres && go generate`
- Migrations use `services/content-api/http/rest/openapi/api.yaml`
  - Migrations are stored in `services/content-api/storage/postgres/migrations`
  - Refer to [these instructions](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md) for a reference on how to write migrations.

## Important links

- [Source code](https://github.com/tadoku/tadoku/tree/main/services/content-api)
- [API Spec](https://github.com/tadoku/tadoku/blob/main/services/content-api/http/rest/openapi/api.yaml)
