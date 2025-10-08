# Implementation Plan for Free Text Tags Feature

## Overview

This feature will replace the existing tag selection functionality with a free text tag input system. Users will be able to create and manage custom tags when submitting logs to Tadoku. The system will support autocomplete based on previously used tags (plus predefined tags as suggestions), allow for multiple tag selection, and ensure accessibility across devices including mobile.

Key architectural decisions:

- Tags are NOT tied to specific activities (simplified from current behavior)
- Maximum 10 tags per log, 50 characters per tag
- Existing predefined tags table has been renamed to `log_default_tags` and will be used as additional suggestions
- Tags are case-insensitive (stored lowercase, displayed as entered)

During implementation, we will use a dual storage approach to maintain backward compatibility with existing features while enabling new functionality:

- Continue using the existing `tags` array in the `logs` table for backward compatibility
- Add a new `log_tags` table for user-created tags functionality
- Write to both storage mechanisms when creating or updating logs
- Gradually transition features to use the new table structure

## Steps

### Step 1: Create API Endpoint for Tag Retrieval

Location:
`services/immersion-api/http/rest/openapi/api.yaml`

Change description:
Add a new endpoint to retrieve a user's previously used tags plus predefined tags from `log_default_tags`. This will support the autocomplete functionality in the frontend.

- Create a GET endpoint at `/users/{userId}/tags` that returns an array of tag strings
- Ensure the endpoint is authenticated and only returns tags for the authenticated user
- Add pagination support to handle potentially large numbers of tags
- Tags should be returned in order of frequency (most used first), with predefined tags appended at the end
- Add prefix filtering to support efficient autocomplete

References:

- `services/immersion-api/http/rest/openapi/api.yaml`
- Existing user endpoints in the API spec (lines 519-596)

Change snippet:

```yaml
paths:
  /users/{userId}/tags:
    get:
      summary: Get user's tags
      description: Returns a list of tags previously used by the user, plus predefined tag suggestions
      operationId: getUserTags
      security:
        - cookieAuth: []
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: prefix
          in: query
          required: false
          description: Filter tags that start with this prefix (case-insensitive)
          schema:
            type: string
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            default: 50
        - name: offset
          in: query
          required: false
          schema:
            type: integer
            default: 0
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
                    description: Array of tag strings (user tags by frequency, then predefined tags)
                  total:
                    type: integer
                    description: Total count of matching tags
        "401":
          $ref: "#/components/responses/Unauthorized"
        "403":
          $ref: "#/components/responses/Forbidden"
```

Notes:

- Tags are stored lowercase in the database
- The prefix filter parameter enables efficient autocomplete functionality
- User's own tags are returned first (sorted by frequency), then predefined tags from `log_default_tags`
- Predefined tags are no longer filtered by activity

### Step 2: Implement Tag Storage and Retrieval in Backend

Location:
`services/immersion-api/storage/postgres/migrations/`

Change description:
Create a database migration to ensure tags are properly stored and can be efficiently queried. Implement a dual storage approach to maintain compatibility with existing features.

- Create a new `log_tags` table to store user-created tags with proper schema
- Keep the existing `tags` array in the `logs` table for backward compatibility
- Create indexes for efficient tag retrieval and frequency counting
- Add SQL queries for tag operations

References:

- `services/immersion-api/storage/postgres/migrations/`
- `services/immersion-api/storage/postgres/queries/`

Notes:

- Tags are stored in lowercase (normalization happens in application code before insert)
- The `user_id` is duplicated from `logs.user_id` for query efficiency
- Maximum 50 characters per tag enforced at database level
- Unique constraint on (log_id, tag) prevents duplicate tags on same log
- The query combines user tags (by frequency) with predefined tags (alphabetically)
- Use `text_pattern_ops` index for efficient prefix matching with LIKE queries
- Transactions must be used when writing to both `logs.tags` array and `log_tags` table

### Step 3: Implement Backend Tag Endpoints

Location:
`services/immersion-api/http/rest/`

Change description:
Implement the handler for the new tag retrieval endpoint.

- Create a new handler file `server_getusertags.go` for GET `/users/{userId}/tags`
- Implement the query service method to fetch tags from `log_tags` and `log_default_tags`
- Ensure proper authorization (users can only access their own tags)

References:

- `services/immersion-api/http/rest/handlers/`
- `services/immersion-api/storage/postgres/queries/`

Notes:

- Follow existing authentication patterns in the codebase
- The query service method should use the SQL queries defined in Step 2

### Step 4: Update Log Submission Endpoints

Location:
`services/immersion-api/http/rest/server_logcreate.go` and related files

Change description:
Enhance the existing log submission to handle free text tags with proper validation and dual-write to both storage mechanisms.

- Update `server_logcreate.go` handler to pass tags to the command service
- Modify `services/immersion-api/domain/command/createlog.go` to validate tags (max 10 tags, max 50 chars each, normalize to lowercase)
- Update repository layer to write tags to both `logs.tags` array and `log_tags` table within same transaction
- Handle tag deduplication (case-insensitive)

References:

- `services/immersion-api/http/rest/server_logcreate.go` (existing handler)
- `services/immersion-api/domain/command/createlog.go` (existing command logic)
- `services/immersion-api/storage/postgres/repository/repo_createlog.go` (repository implementation)
- `services/immersion-api/storage/postgres/queries/logs.sql`

Notes:

- Tags are normalized to lowercase and deduplicated before storage
- Maximum 10 tags per log, 50 characters per tag (enforced in application and database)
- Empty/whitespace-only tags are filtered out
- Both storage mechanisms must be updated within the same database transaction for consistency
- If tag insertion fails, the entire log creation is rolled back

### Step 5: Create Data Migration for Existing Tags

Location:
`services/immersion-api/storage/postgres/migrations/`

Change description:
Create a migration script to copy existing tags from the array in the `logs` table to the new `log_tags` table. This populates the new table with historical data.

- Read all existing non-deleted logs with tags
- Extract tags from the array and insert them into the `log_tags` table
- Ensure tags are converted to lowercase during migration
- Handle duplicates gracefully with ON CONFLICT

References:

- `services/immersion-api/storage/postgres/migrations/`

Notes:

- Migration runs as part of the schema migration, no separate batch processing needed
- Filters out empty/whitespace tags and tags exceeding 50 characters
- ON CONFLICT clause prevents duplicate entries if migration is re-run
- Only migrates non-deleted logs to avoid unnecessary data
- The migration is idempotent and safe to run multiple times

### Step 6: Create Tag Input Component

Location:
`frontend/packages/ui/components/Form/`

Change description:
Create a NEW reusable tag input component for free-text tag entry with autocomplete. This will be separate from the existing `AutocompleteMultiInput` component.

Component requirements:

- Accept free-text input (not limited to predefined options)
- Show autocomplete suggestions from user's history + predefined tags
- Support multiple tag selection (max 10)
- Visual tags with remove buttons
- Enforce 50 character limit per tag
- Accessibility (keyboard navigation, screen reader support)
- Mobile responsiveness
- Case-insensitive tag matching

References:

- `frontend/packages/ui/components/Form/AutocompleteMultiInput.tsx` (for reference, but don't modify)
- `frontend/apps/styleguide/pages/forms.tsx` (add example to styleguide)

Notes:

- Component should integrate with react-hook-form (like other Form components)
- Use downshift for proper ARIA attributes and keyboard navigation
- Style consistently with existing UI components in the package
- Tags are normalized to lowercase when added
- Character counter shows remaining space
- Add to styleguide with interactive examples
- Test with keyboard-only navigation and screen readers

### Step 7: Implement Tag Fetching Hook and API Client

Location:
`frontend/apps/webv2/app/immersion/api.ts`

Change description:
Add API client function and React hook to fetch user tags from the new endpoint with autocomplete support.

- Add API endpoint and Zod schema to `api.ts` (similar to existing immersion API patterns)
- Create a React Query hook for fetching tags with prefix filtering
- Support debounced input for autocomplete
- Cache results appropriately

References:

- `frontend/apps/webv2/app/immersion/api.ts` (existing API patterns around line 687-721)
- Existing hooks like `useLogConfigurationOptions`

Notes:

- Debouncing the input prevents excessive API calls during typing
- The hook respects React Query caching to minimize network requests
- Empty prefix returns user's most frequently used tags plus predefined tags
- With prefix, only matching tags are returned
- Consider implementing a debounce hook if one doesn't exist

### Step 8: Update Log Form to Use New Tag Component

Location:
`frontend/apps/webv2/app/immersion/NewLogForm/`

Change description:
Replace the existing tag selection with the new free text tag input in the log submission form.

- Update `Form.tsx` to use the new `TagInput` component instead of `AutocompleteMultiInput`
- Remove activity-based tag filtering from `domain.tsx`
- Update `NewLogFormSchema` to accept string arrays instead of Tag objects
- Update `NewLogAPISchema` to send tags as strings (already correct, just need to change source type)
- Integrate with `useUserTags` hook for autocomplete

References:

- `frontend/apps/webv2/app/immersion/NewLogForm/Form.tsx` (around line 201-213)
- `frontend/apps/webv2/app/immersion/NewLogForm/domain.tsx` (schema definition around line 33, filterTags function around line 122)
- `frontend/apps/webv2/app/immersion/api.ts`

Notes:

- Tags are now simple strings, not objects with id and activity_id
- No activity-based filtering needed
- The old Tag type and filterTags function can be removed
- Existing logs will display properly since the API returns tags as strings
- Debounce hook may need to be implemented if it doesn't exist

### Step 9: Update Log Display Components

Location:
`frontend/apps/webv2/pages/logs/` and related display components

Change description:
Verify and update log display components to properly show free text tags. The API already returns tags as strings, so display components should work without major changes.

- Verify log detail page (`pages/logs/[id].tsx`) displays tags correctly
- Verify log list views display tags properly
- Ensure consistent tag styling across all views
- Tags should be displayed as simple badges/pills

References:

- `frontend/apps/webv2/pages/logs/[id].tsx` (log detail page)
- `frontend/apps/webv2/app/immersion/` (log list components)
- API response schema in `api.ts` (Log type already has `tags: string[]`)

Notes:

- The Log API response already returns `tags` as `string[]` (see OpenAPI schema line 1256-1259)
- Display components should already work since they consume string arrays
- Main task is to verify styling is consistent and readable
- Consider adding hover state or tooltip for long tag names
- Future enhancement: add clickable tags for filtering logs by tag (not in this iteration)

### Step 10: Update LogConfigurationOptions Endpoint

Location:
`services/immersion-api/http/rest/server_loggetconfigurations.go` and OpenAPI spec

Change description:
Update or deprecate the log configuration options endpoint to reflect that tags are no longer activity-specific predefined options.

- Update OpenAPI schema: tags in `LogConfigurationOptions` should reference `log_default_tags` (not activity-filtered)
- Update backend handler to return all predefined tags (remove activity filtering)
- Consider if `Tag` schema with `id` and `log_activity_id` should be deprecated since tags are now just strings
- Update frontend to handle configuration options without activity-filtered tags

References:

- `services/immersion-api/http/rest/openapi/api.yaml` (lines 842-847, 879-902)
- `services/immersion-api/http/rest/server_loggetconfigurations.go`
- `services/immersion-api/storage/postgres/queries/tags.sql`
- `frontend/apps/webv2/app/immersion/api.ts` (LogConfigurationOptions usage)

Notes:

- This is primarily cleanup since new form doesn't use predefined tags from this endpoint
- The endpoint can still return predefined tags for backward compatibility
- Activity-based tag filtering is completely removed
- Future consideration: deprecate this entire part of the configuration endpoint once frontend migration is complete

### Step 11: Testing and Validation

Location:
Various test files across frontend and backend

Change description:
Implement comprehensive testing for the new tag functionality to ensure reliability and accessibility.

Backend testing:

- Test tag validation (max 10 tags, max 50 chars, lowercase normalization)
- Test dual-write transaction consistency (both `logs.tags` array and `log_tags` table)
- Test GET `/users/{userId}/tags` endpoint with various scenarios (empty, with prefix, pagination)
- Test tag frequency ordering and predefined tag merging
- Test tag deduplication logic

Frontend testing:

- Unit tests for TagInput component (add, remove, keyboard navigation)
- Test tag input validation and character limit
- Test autocomplete filtering and display
- Integration tests for log form submission with tags
- Accessibility testing (keyboard navigation, screen reader compatibility)
- Mobile responsiveness testing

References:

- Backend test patterns in `services/immersion-api/domain/command/`
- Frontend component test examples (if they exist)
- `frontend/apps/styleguide` for manual testing

Testing checklist:

- [ ] Keyboard navigation (Enter to add, Backspace to remove, Tab to navigate)
- [ ] Screen reader announces tags correctly
- [ ] Mobile touch interactions work smoothly
- [ ] Tags persist correctly after page refresh
- [ ] Special characters and edge cases handled properly
- [ ] Database transaction rollback on error
- [ ] API rate limiting (if implemented)

Notes:

- Use accessibility tools like Axe or WAVE for automated testing
- Manual testing with screen readers (VoiceOver, NVDA)
- Test on various mobile devices (iOS Safari, Android Chrome)
- Test with different tag lengths and quantities
- Verify migration script works correctly with existing data

### Step 12: Clean Up Legacy Tag Implementation

Location:
Various backend and frontend files

Change description:
After the new tag system has been running successfully in production for a reasonable period (e.g., 2-4 weeks), clean up the legacy dual-write implementation.

Phase 1 - Stop dual-writes (requires careful coordination):

- Update `repo_createlog.go` to write ONLY to `log_tags` table
- Remove writes to `logs.tags` array
- Keep reads from both for a transition period

Phase 2 - Update all queries to use `log_tags` table:

- Update any remaining queries that read from `logs.tags` array
- Ensure all log retrieval uses `log_tags` table via JOIN

Phase 3 - Database cleanup (optional, can be deferred):

- Create migration to drop the `tags` column from `logs` table
- Update Go models to remove Tags field from Log struct
- Update OpenAPI schema to remove tags from Log response (if sourced from table instead)

References:

- `services/immersion-api/storage/postgres/repository/repo_createlog.go`
- `services/immersion-api/storage/postgres/queries/logs.sql`
- `services/immersion-api/storage/postgres/models.go`
- `services/immersion-api/http/rest/openapi/api.yaml`

Notes:

- **CRITICAL**: Only proceed after thorough production testing
- Consider keeping the `logs.tags` column for 3-6 months even after stopping writes (for easy rollback)
- Monitor for any issues after each phase before proceeding to next
- Document the change and update any README files
- This step has the highest risk, so take it slowly
- Keep backups before running any destructive migrations

Recommended timeline:

- Phase 1: After 2-4 weeks of successful dual-write operation
- Phase 2: After 1-2 weeks of Phase 1 running smoothly
- Phase 3: After 1-3 months of Phase 2 running smoothly (or defer indefinitely)

## Minimum Context

Key files to understand before implementing:

Backend:

- `services/immersion-api/http/rest/openapi/api.yaml` - API specification
- `services/immersion-api/storage/postgres/migrations/0001_init.up.sql` - Current schema
- `services/immersion-api/storage/postgres/queries/logs.sql` - Log queries
- `services/immersion-api/storage/postgres/queries/tags.sql` - Tag queries (to be updated)
- `services/immersion-api/http/rest/server_logcreate.go` - Log creation handler
- `services/immersion-api/domain/command/createlog.go` - Log creation logic
- `services/immersion-api/storage/postgres/repository/repo_createlog.go` - Repository layer

Frontend:

- `frontend/packages/ui/components/Form/AutocompleteMultiInput.tsx` - Reference component
- `frontend/apps/webv2/app/immersion/api.ts` - API client and hooks
- `frontend/apps/webv2/app/immersion/NewLogForm/Form.tsx` - Log form component
- `frontend/apps/webv2/app/immersion/NewLogForm/domain.tsx` - Form schema and validation
- `frontend/apps/webv2/pages/logs/[id].tsx` - Log detail page

Database:

- Table `logs` has `tags varchar(50)[]` column (current implementation)
- Table `log_default_tags` contains predefined tag suggestions (renamed from `log_tags`)
- New table `log_tags` will be created for user tags
