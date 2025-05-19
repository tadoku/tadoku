# [004] Automating contest scheduling

- Status: accepted
- Author: @antonve
- Date: 2025-05-16

## Terminology

- Official contest: a contest that counts for the official leaderboard. Scheduled by the maintainers of Tadoku several times a year.

## Context

Scheduling official contests is a manual process that needs to be performed between the end of the previous contest and start of the next contest. Ideally, this should happen 1–2 weeks before the new contest starts, for the following reasons:

- We want to show the leaderboard of the previous contest on the [Latest leaderboard](https://tadoku.app/leaderboard/latest/1) page so people can check the final leaderboard after a contest has finished.
- Early visibility encourages more sign-ups; people are less likely to join after a contest has started.

We want to schedule these contests automatically, without requiring any manual intervention. We also want to be notified through an alert on Discord in case there is a failure.

## Considered options

### Solution 1: Cron Job in Kubernetes

Similar to how we implemented `postgres-backup`, we can create a bash script to schedule an official contest. This can then be configured to run on a schedule as a Kubernetes CronJob.

We already have monitoring configured for the backup job. We can use the same mechanism for monitoring the cron job.

### Solution 2: PostgreSQL with pg_cron

PostgreSQL supports running SQL on a schedule with the `pg_cron`. This is already installed by default when using the [zalando/postgres-operator](https://github.com/zalando/postgres-operator) (which we use for Tadoku). One advantage of doing so is that we can keep it all in the database without having any external dependencies.

We'd encapsulate the scheduling logic in a database function, with sufficient validation to prevent invalid contest creation. However, this means there is business logic in the database which needs to be maintained.

Monitoring would have to be implemented outside of the database. We're thinking of building a Discord bot to help with operations, and this component could possibly implement this.

### Solution 3: Background job in immersion-api

The API responsible for managing contests `immersion-api` can check if a contest needs to be scheduled in a background job. One limitation with this approach is that we need to ensure we only have one instance trying to schedule a contest at a time. We'd need some mechanism to ensure this, which would make this approach more complex than the other solutions discussed above.

## Decision

| Criteria                   | Solution 1: Cron Job in Kubernetes                                                                        | Solution 2: PostgreSQL with pg_cron                                      | Solution 3: Background job in immersion-api                                                  |
| -------------------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------- |
| Initial effort             | Low.                                                                                                      | Low.                                                                     | High. Needs some kind of mechanism to prevent duplicated contests (such as leader election). |
| Maintenance effort         | Medium. We need to ensure the CronJob persists in case we ever migrate to a different Kubernetes cluster. | Medium. Need to maintain business logic in the database.                 | High. Needs to keep working during deployments.                                              |
| Monitoring ease            | Easy.                                                                                                     | Medium. Needs an external monitor.                                       | Medium. Need to integrate Discord API within the service.                                    |
| **Overall recommendation** | **2nd choice** - simple and effective, but adds a new component.                                          | **1st choice** - easy to implement without introducing new dependencies. | **Not recommended** - adds unnecessary complexity.                                           |

## Outcome

I've [implemented](https://github.com/tadoku/tadoku/commit/5a15fba023d09f09c4899b8d0fcea9b0cb2b0102) it using solution 2. There were some challenges which I wasn't aware of previously:

1. `pg_cron` can only be installed in one database (usually the default `postgres` database).
2. `pg_cron` creates a new connection for running the job. This means we need to configure the job with the right credentials in order to run. The predefined scheduling functions don’t support socket-based connections, which required us to insert the job into the database manually.
3. The database is slightly different in production and development. I've opted to create a staging database on the k8s cluster which mimics the production cluster (minus the actual production data). This was used to develop the function and test scheduling functionality.
4. Since scheduling is done in a different database, we had to manually run them instead of including them in the migration files.

In retrospect, solution 1 may have been simpler to implement. However, it's not worth the time migrating over now that we have solution 2 working.

Monitoring is pending the creation of the Operations Discord bot.
