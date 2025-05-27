# Implementation Plan for Free Text Tags Feature

## Overview

This feature will replace the existing tag selection functionality with a free text tag input system. Users will be able to create and manage custom tags when submitting logs to Tadoku. The system will support autocomplete based on previously used tags, allow for multiple tag selection, and ensure accessibility across devices including mobile.

During implementation, we will use a dual storage approach to maintain backward compatibility with existing features while enabling new functionality:

- Continue using the existing `tags` array in the `logs` table for backward compatibility
- Add a new `log_tags` table for the enhanced tag functionality
- Write to both storage mechanisms when creating or updating logs
- Gradually transition features to use the new table structure

## Steps

### Step 1: Create API Endpoint for Tag Retrieval [COMPLETE]

Location:
`services/immersion-api/http/rest/openapi/api.yaml`

Change description:
Add a new endpoint to retrieve a user's previously used tags. This will support the autocomplete functionality in the frontend.

- Create a GET endpoint at `/users/{userId}/tags` that returns an array of tag strings
- Ensure the endpoint is authenticated and only returns tags for the authenticated user
- Return only the top 20 most frequently used tags
- Tags should be returned in order of frequency (most used first)
- Add prefix filtering to support efficient autocomplete

References:

- `services/immersion-api/http/rest/openapi/api.yaml`
- Existing user endpoints in the API spec

Change snippet:

```yaml
paths:
  /users/{userId}/tags:
    get:
      summary: Get user's tags
      description: Returns a list of the top 20 most frequently used tags by the user that match the given prefix
      operationId: getUserTags
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
        - name: prefix
          in: query
          required: false
          description: Filter tags that start with this prefix
          schema:
            type: string
      responses:
        "200":
          description: List of user tags
          content:
            application/json:
              schema:
                type: object
                properties:
                  tags:
                    type: array
                    items:
                      type: string
        "401":
          description: Unauthorized
        "403":
          description: Forbidden
```

Notes:

- The endpoint should return tags in their original case as stored in the database
- Tags will be lowercased when written to the database, not when retrieved
- The prefix filter parameter will enable efficient autocomplete functionality

### Step 2: Create Copy of Existing `log_tags` Table as `log_tags_legacy` [NOT STARTED]

Location:
`services/immersion-api/storage/postgres/migrations/`

Change description:
Before implementing the new tag storage system, create a copy of the existing `log_tags` table as `log_tags_legacy`. This approach supports gradual feature rollout by preserving the existing table while we develop and test the new implementation.

- Create a migration to copy the existing `log_tags` table structure and data to `log_tags_legacy`
- Create necessary indexes and constraints on the new table
- Keep the original `log_tags` table intact to maintain existing functionality during development

References:

- `services/immersion-api/storage/postgres/migrations/`
- `services/immersion-api/storage/postgres/queries/`
- `services/immersion-api/http/rest/handlers/`

Change snippet:

```sql
-- Example migration
create table if not exists log_tags_legacy as table log_tags;

-- Recreate primary key
alter table log_tags_legacy add primary key (log_id, tag);

-- Recreate indexes (adjust based on actual existing indexes)
create index if not exists idx_log_tags_legacy_tag on log_tags_legacy(tag);
create index if not exists idx_log_tags_legacy_user on log_tags_legacy(user_id, tag);

-- Recreate foreign key constraints
alter table log_tags_legacy add constraint log_tags_legacy_log_id_fkey
  foreign key (log_id) references logs(id) on delete cascade;
```

Notes:

- This approach maintains the existing `log_tags` table for current functionality
- The copied table allows us to safely develop and test the new implementation
- Enables gradual rollout of features without breaking existing functionality
- Consider adding a comment to the legacy table to document its purpose for future maintenance
- Include appropriate logging and monitoring during the migration process

### Step 3: Create Migration for New `log_tags` Table [NOT STARTED]

Location:
`services/immersion-api/storage/postgres/migrations/`

Change description:
Create a database migration to add the new `log_tags` table and necessary indexes for efficient querying.

- Create a new `log_tags` table to store tags
- Create indexes for efficient tag retrieval

References:

- `services/immersion-api/storage/postgres/migrations/`

Change snippet:

```sql
-- Example migration (actual implementation may vary based on current schema)
create table if not exists log_tags (
  log_id uuid references logs(id) on delete cascade,
  tag text not null,
  primary key (log_id, tag)
);

create index idx_log_tags_tag on log_tags(tag);
create index idx_log_tags_user on log_tags(user_id, tag);
```

Notes:

- Ensure proper foreign key constraints are in place
- Create appropriate indexes for performance optimization
- Consider adding constraints for tag length or format if needed

### Step 4: Create Migration to Copy Existing Tag Data [NOT STARTED]

Location:
`services/immersion-api/storage/postgres/migrations/`

Change description:
Create a migration script to copy existing tags from the array in the `logs` table to the new `log_tags` table.

- Implement a migration that reads all existing logs
- Extract tags from the array and insert them into the `log_tags` table
- Ensure tags are converted to lowercase during migration
- Process in batches to minimize database load

References:

- `services/immersion-api/storage/postgres/migrations/`

Change snippet:

```sql
-- Example migration to copy existing tags (implement in batches for production)
insert into log_tags (log_id, user_id, tag)
select l.id, l.user_id, lower(t) as tag
from logs l, unnest(l.tags) as t
on conflict do nothing;
```

Notes:

- Consider running this migration during off-peak hours
- Implement batching for large datasets
- Add monitoring to track migration progress
- Have a rollback plan in case of issues

### Step 5: Update Tag-Related Queries to Use Both Storage Mechanisms [NOT STARTED]

Location:
`services/immersion-api/storage/postgres/queries/`
`services/immersion-api/http/rest/handlers/`

Change description:
Modify existing tag-related queries to write to both the array in the `logs` table and the new `log_tags` table.

- Update tag create/update/delete operations to maintain both storage systems
- Ensure transactions are used for data consistency
- Implement error handling for cases where one storage mechanism fails

References:

- `services/immersion-api/storage/postgres/queries/`
- `services/immersion-api/http/rest/handlers/`

Notes:

- Use database transactions to ensure atomicity
- Consider performance implications of writing to both storage mechanisms
- Implement proper error handling and logging
- Consider gradually transitioning read operations to use the new table

### Step 6: Implement Backend Tag Endpoints [NOT STARTED]

Location:
`services/immersion-api/http/rest/handlers/`

Change description:
Implement the handler for the new tag endpoint and update existing log submission endpoints to handle free text tags.

- Create handler for GET `/users/{userId}/tags` using the new `log_tags` table
- Update log submission/update handlers to process free text tags and write to both storage mechanisms

References:

- `services/immersion-api/http/rest/handlers/`
- `services/immersion-api/storage/postgres/queries/`

Notes:

- Consider adding rate limiting for tag-related endpoints
- Ensure transactions are used to maintain consistency between both tag storage mechanisms

### Step 7: Update Log Submission Endpoints [NOT STARTED]

Location:
`services/immersion-api/http/rest/handlers/logs.go`

Change description:
Enhance the existing log submission endpoints to properly handle the new free text tags format.

- Update log creation and update endpoints to normalize tags to lowercase
- Add validation to prevent excessively long tags or too many tags per log
- Implement writing to both the array in the `logs` table and the new `log_tags` table
- Ensure proper transaction handling for data consistency

References:

- `services/immersion-api/http/rest/handlers/logs.go`
- `services/immersion-api/storage/postgres/queries/logs.sql`

Notes:

- Consider implementing a maximum length for tags (e.g., 50 characters)
- Consider implementing a maximum number of tags per log (e.g., 10 tags)
- Add validation to reject tags with special characters if needed
- Ensure both storage methods are updated within the same transaction

### Step 8: Create Data Migration for Existing Tags [NOT STARTED]

Location:
`services/immersion-api/storage/postgres/migrations/`

Change description:
Create a migration script to copy existing tags from the array in the `logs` table to the new `log_tags` table.

- Implement a migration that reads all existing logs
- Extract tags from the array and insert them into the `log_tags` table
- Ensure tags are converted to lowercase during migration
- Process in batches to minimize database load

References:

- `services/immersion-api/storage/postgres/migrations/`

Change snippet:

```sql
-- Example migration to copy existing tags (implement in batches for production)
insert into log_tags (log_id, user_id, tag)
select l.id, l.user_id, lower(t) as tag
from logs l, unnest(l.tags) as t
on conflict do nothing;
```

Notes:

- Consider running this migration during off-peak hours
- Implement batching for large datasets
- Add monitoring to track migration progress
- Have a rollback plan in case of issues

### Step 9: Create Tag Input Component [NOT STARTED]

Location:
`frontend/packages/ui/src/components/`

Change description:
Create a reusable tag input component that supports the required functionality:

- Multiple tag input
- Autocomplete from previously used tags
- Visual distinction for selected tags
- Tag removal
- Accessibility features
- Mobile responsiveness

References:

- `frontend/packages/ui/src/components/`
- `frontend/apps/styleguide`

Change snippet:

```tsx
// Example component structure (actual implementation will be more detailed)
import React, { useState, useEffect } from "react";
import { useCombobox, useMultipleSelection } from "downshift";

export type TagInputProps = {
  initialTags?: string[];
  onChange: (tags: string[]) => void;
  placeholder?: string;
  disabled?: boolean;
  maxTags?: number;
  className?: string;
  suggestedTags?: string[];
  isLoading?: boolean;
  onInputChange?: (inputValue: string) => void;
};

export const TagInput: React.FC<TagInputProps> = ({
  initialTags = [],
  onChange,
  placeholder = "Add tags...",
  disabled = false,
  maxTags = 10,
  className = "",
  suggestedTags = [],
  isLoading = false,
  onInputChange,
}) => {
  // Implementation details will go here
  // Will use downshift for accessibility and keyboard navigation
  // Will handle mobile-specific interactions
  // Will implement tag selection, removal, and input focus management
};
```

Notes:

- Use `downshift` or a similar library to ensure accessibility
- Ensure the component works well with screen readers
- Test thoroughly on mobile devices
- Consider keyboard navigation and focus management

### Step 10: Implement Tag Fetching Hook [NOT STARTED]

Location:
`frontend/apps/webv2/src/hooks/`

Change description:
Create a React hook to fetch user tags from the new API endpoint.

- Implement a hook that fetches tags from the new endpoint
- Include caching and error handling
- Support filtering/searching tags for autocomplete
- Integrate with React Query for data fetching

References:

- `frontend/apps/webv2/src/hooks/`
- `frontend/apps/webv2/src/api/`

Change snippet:

```tsx
// Example hook implementation
import { useQuery } from "@tanstack/react-query";
import { z } from "zod";
import { api } from "../api/client";

const tagsResponseSchema = z.object({
  tags: z.array(z.string()),
  total: z.number(),
});

export function useUserTags(userId: string, search?: string) {
  return useQuery({
    queryKey: ["userTags", userId, search],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (search) params.append("search", search);

      const response = await api.get(`/users/${userId}/tags?${params}`);
      return tagsResponseSchema.parse(response.data);
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
```

Notes:

- Consider debouncing search queries to prevent excessive API calls
- Implement proper error handling and loading states

### Step 11: Update Log Form to Use New Tag Component [NOT STARTED]

Location:
`frontend/apps/webv2/src/components/logs/`

Change description:
Replace the existing tag selection component with the new free text tag input component in the log submission form.

- Update the log form to use the new TagInput component
- Connect the component to the tag fetching hook
- Ensure form validation handles the new tag format
- Update the form submission logic to send tags in the required format

References:

- `frontend/apps/webv2/src/components/logs/`
- `frontend/apps/webv2/src/hooks/`

Notes:

- Ensure backward compatibility with existing logs
- Consider adding a migration path for users with existing logs

### Step 12: Update Log Display Components [NOT STARTED]

Location:
`frontend/apps/webv2/src/components/logs/`

Change description:
Update components that display logs to properly show the new free text tags.

- Update log list and detail views to display free text tags
- Ensure consistent styling between input and display components

References:

- `frontend/apps/webv2/src/components/logs/`

Notes:

- Consider adding a way to search/filter logs by tag
- Ensure tag display is consistent across all views

### Step 13: Update Tag-Dependent Features to Use New Table [NOT STARTED]

Location:
Various backend and frontend files

Change description:
Identify and update features that depend on tags to use the new `log_tags` table instead of the array in the `logs` table.

- Update tag filtering functionality to use the new table
- Update tag statistics and reporting features
- Ensure performance is maintained or improved with the new structure

References:

- `services/immersion-api/http/rest/handlers/`
- `services/immersion-api/storage/postgres/queries/`
- `frontend/apps/webv2/src/components/logs/`

Notes:

- Prioritize features that would benefit most from the new structure (like prefix search)
- Maintain backward compatibility during the transition
- Add feature flags if needed to control which implementation is used

### Step 14: Testing and Validation [NOT STARTED]

Change description:
Implement comprehensive testing for both frontend and backend components.

- Add unit tests for the TagInput component
- Test accessibility using automated tools and manual testing
- Test mobile responsiveness across different device sizes

References:

- Frontend and backend test directories

Notes:

- Use tools like Axe for accessibility testing
- Test with screen readers
- Test with keyboard-only navigation
- Test on various mobile devices and screen sizes

### Step 15: Clean Up Legacy Tag Implementation [NOT STARTED]

Location:
Various backend and frontend files

Change description:
Once all features have been migrated to use the new `log_tags` table, clean up the legacy implementation.

- Remove code that writes to the tags array in the `logs` table
- Update all remaining queries to use only the `log_tags` table
- Consider a database migration to remove the tags column from the `logs` table (optional, can be done later)

References:

- `services/immersion-api/http/rest/handlers/`
- `services/immersion-api/storage/postgres/queries/`
- `frontend/apps/webv2/src/components/logs/`

Notes:

- This step should only be done after thorough testing
- Consider keeping the array column for a period even after code is updated (for easy rollback)
- Update documentation to reflect the new tag implementation

## Minimum Context

- `services/immersion-api/http/rest/openapi/api.yaml`
- `services/immersion-api/storage/postgres/migrations/`
- `services/immersion-api/storage/postgres/queries/`
- `services/immersion-api/http/rest/handlers/`
- `frontend/packages/ui/src/components/`
- `frontend/apps/webv2/src/hooks/`
- `frontend/apps/webv2/src/api/`
- `frontend/apps/webv2/src/components/logs/`

## Status Legend

[COMPLETE] - Task is finished and verified
[IN PROGRESS] - Task is currently being worked on
[NOT STARTED] - Task has not been started yet
