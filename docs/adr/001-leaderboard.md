# [001] Leaderboard Improvements

* Status: proposed
* Author: @antonve
* Date: 2022-04-23

## Context

As of now the leaderboard has limitations that prevents us from delivering a better UX. We're limited to just one variation of a leaderboard, without the ability to filter or change the start and/or end date of a leaderboard.
We'd like to introduce global and yearly leaderboards, together with the ability to filter on language and/or media. The current design where we precalculate the rankings does not work given these constraints. 
We also have no way of knowing a user's position in the leaderboard without fetching the complete leaderboard. It would be nice if we could show the current position on a user's profile page.
Global and yearly leaderboards could get fairly large, so ideally we would be able to paginate the leaderboard.

### Use cases

![Overview of the use cases](assets/001/use-cases.png)

## Considered options

### PostgreSQL without precalculations

We could get rid of the materialized leaderboards in the database all-together. This allows us to add filters, or generate leaderboards over different timespans.

An example query to fetch a leaderboard could look like this:

```sql
with ranks as (
  select user_id, sum(weighted_score) as score
  from contest_logs
  where
    -- we can change these conditions on the fly to get a different variation
    contest_id = 9 and
    language_code = 'jpn'
  group by user_id
)

select r.*, u.display_name
from ranks as r
inner join users as u on (u.id = r.user_id)
order by r.score desc
```

#### Pros

* Simple design
* Flexible enough so we can add more filters if we wish to do so

#### Cons

* Unable to get a user's position in a leaderboard without fetching the complete leaderboard
* Pagination with a design like this is hard on the database (for global/yearly leaderboards)

## Decision

## Outcome
