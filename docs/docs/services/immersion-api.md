---
sidebar_position: 1
title: immersion-api
---

# Immersion API

This is the main API of the Tadoku website. It contains the functionality to:

- Log activities consuming language
- Create and participate in contests
- Leaderboards for contests
- Visualize logged activities

All core functionality should live in this service.

## Architecture

- This service is written in Golang.
- OpenAPI REST API over HTTP
  - Spec can be found at `services/immersion-api/http/rest/openapi/api.yaml`
- Uses CQRS (Command Query Responsibility Segregation) to split data read/writes
- As of now, there's only one service for query and command. This is because the boundaries for different services was not yet clear (we'll want to improve this).
- Data is stored in a PostgreSQL database
- sqlc is used to generate Golang glue code for raw sql queries.
  - New queries should be written in `services/immersion-api/storage/postgres/queries`
  - Run the following command to generate code for the query: `cd services/immersion-api/storage/postgres && go generate`
- After changing the OpenAPI spec, regenerate the Go server code:
  ```
  cd services/immersion-api/http/rest/openapi && go generate
  ```
- Migrations use `services/immersion-api/http/rest/openapi/api.yaml`
  - Migrations are stored in `services/immersion-api/storage/postgres/migrations`
  - Refer to [these instructions](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md) for a reference on how to write migrations.

## Important links

- [Source code](https://github.com/tadoku/tadoku/tree/main/services/immersion-api)
- [API Spec](https://github.com/tadoku/tadoku/blob/main/services/immersion-api/http/rest/openapi/api.yaml)
