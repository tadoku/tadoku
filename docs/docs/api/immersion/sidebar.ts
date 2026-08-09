import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/immersion/immersion-api",
    },
    {
      type: "category",
      label: "contests",
      items: [
        {
          type: "doc",
          id: "api/immersion/contest-create",
          label: "Creates a new contest",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/immersion/contest-list",
          label: "Lists all the contests, paginated",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-create-permission-check",
          label: "Check if user has permission to create a new contest",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-find-by-id",
          label: "Fetches a contest by id",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-find-latest-official",
          label: "Fetches the latest official contest",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-find-registration",
          label: "Fetches a contest registration if it exists",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-registration-upsert",
          label: "Creates or updates a registration for a contest",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/immersion/contest-fetch-leaderboard",
          label: "Fetches the leaderboard for a contest",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-fetch-summary",
          label: "Fetches the summary for a contest",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-list-logs",
          label: "Lists the logs attached to a contest",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-moderation-detach-log",
          label: "Detaches a log from a contest (moderation action)",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/immersion/contest-find-ongoing-registrations",
          label: "Fetches all the ongoing contest registrations of the logged in user, always in a single page",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-get-configurations",
          label: "Fetches the configuration options for a new contest",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "contests_profile",
      items: [
        {
          type: "doc",
          id: "api/immersion/contest-profile-fetch-scores",
          label: "Fetches the scores of a user profile in a contest",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/contest-profile-fetch-activity",
          label: "Fetches the activity of a user profile in a contest",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "logs",
      items: [
        {
          type: "doc",
          id: "api/immersion/log-create",
          label: "Submits a new log",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/immersion/score-preview",
          label: "Previews platform and contest scores without creating a log",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/immersion/log-find-by-id",
          label: "Fetches a log by id",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/log-delete-by-id",
          label: "Deletes a log by id",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/immersion/log-update",
          label: "Updates an existing log",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "api/immersion/log-contest-registration-update",
          label: "Updates the contest registrations for a log",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "api/immersion/log-get-configurations",
          label: "Fetches the configuration options for a log",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/log-tag-suggestions",
          label: "Fetches tag suggestions for autocomplete",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "profile",
      items: [
        {
          type: "doc",
          id: "api/immersion/profile-find-by-user-id",
          label: "Fetches a profile of a user",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/profile-yearly-activity-by-user-id",
          label: "Fetches a activity summary of a user for a given year",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/profile-yearly-scores-by-user-id",
          label: "Fetches the scores of a user for a given year",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/profile-list-logs",
          label: "Lists the logs of a user",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/profile-yearly-contest-registrations-by-user-id",
          label: "Fetches the contest registrations of a user for a given year",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/profile-yearly-activity-split-by-user-id",
          label: "Fetches a activity split summary of a user for a given year",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "leaderboard",
      items: [
        {
          type: "doc",
          id: "api/immersion/fetch-leaderboard-for-year",
          label: "Fetches the leaderboard for a given year",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/fetch-leaderboard-global",
          label: "Fetches the global leaderboard",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "scoring",
      items: [
        {
          type: "doc",
          id: "api/immersion/scoring-rule-set-list-platform",
          label: "Lists platform scoring rule-set versions",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/scoring-rule-set-create-platform",
          label: "Creates a draft platform scoring rule-set version",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/immersion/scoring-rule-set-list-contest",
          label: "Lists scoring rule-set versions owned by a contest",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/scoring-rule-set-create-contest",
          label: "Creates a draft contest scoring rule-set version",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/immersion/scoring-rule-set-publish",
          label: "Publishes an immutable scoring rule-set version",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/immersion/scoring-rule-set-activate",
          label: "Activates a published scoring rule-set version",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "metadata",
      items: [
        {
          type: "doc",
          id: "api/immersion/ping",
          label: "Checks if service is responsive",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "admin",
      items: [
        {
          type: "doc",
          id: "api/immersion/language-list",
          label: "Lists all languages (admin only)",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/immersion/language-create",
          label: "Creates a new language (admin only)",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/immersion/language-update",
          label: "Updates an existing language (admin only)",
          className: "api-method put",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
