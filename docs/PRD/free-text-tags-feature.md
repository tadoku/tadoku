# Free Text Tags Feature

## Overview

Replace the existing activity-filtered tag dropdown with a free-text tag input. Users can type custom tags with autocomplete suggestions from their history and predefined defaults.

**Constraints:**
- Max 10 tags per log, 50 characters per tag
- Case-insensitive (stored lowercase)
- Tags are NOT tied to activities

**Storage:** New `log_tags` table stores user tags (one row per tag per log). The `logs.tags` array will be migrated and removed.

---

## Tasks

### 1. Database: Create log_tags table and migrate data

**Location:** `services/immersion-api/storage/postgres/migrations/`

Single migration that:
- Creates new `log_tags` table with indexes
- Copies existing tags from `logs.tags` array (normalized to lowercase, skip invalid)
- Drops the `tags` column from `logs` table

```sql
CREATE TABLE log_tags (
  log_id UUID NOT NULL REFERENCES logs(id) ON DELETE CASCADE,
  user_id UUID NOT NULL,
  tag VARCHAR(50) NOT NULL,
  PRIMARY KEY (log_id, tag)
);
CREATE INDEX idx_log_tags_user_tag ON log_tags (user_id, tag);
CREATE INDEX idx_log_tags_user_prefix ON log_tags (user_id, tag text_pattern_ops);
```

### 2. Backend: Add tag validation to log creation

**Location:** `services/immersion-api/domain/command/createlog.go`

Add validation before creating a log:
- Max 10 tags
- Max 50 characters per tag
- Normalize to lowercase
- Deduplicate
- Strip empty/whitespace-only tags

### 3. Backend: Update log creation to write to log_tags table

**Location:** `services/immersion-api/storage/postgres/repository/repo_createlog.go`

Insert tags into `log_tags` table instead of `logs.tags` array.

### 4. Backend: Update log queries to read from log_tags table

**Location:** `services/immersion-api/storage/postgres/queries/logs.sql`

Update log retrieval queries to JOIN with `log_tags` and aggregate tags into array for API response.

### 5. Backend: Implement endpoint to fetch user's tags for autocomplete

**Location:** `services/immersion-api/http/rest/server_getusertags.go` (new)

Implement GET `/users/{userId}/tags` (already in OpenAPI spec):
- Query `log_tags` table for user's tags with frequency count
- Support `prefix` query param for filtering
- Append predefined tags from `log_default_tags`
- Return most frequently used first

### 6. Frontend: Create TagInput component

**Location:** `frontend/packages/ui/components/Form/TagInput.tsx`

New component for free-text tag entry:
- Text input that creates tags on Enter/comma
- Autocomplete dropdown from suggestions
- Visual tag pills with remove button
- Character counter (50 max)
- Keyboard accessible (Enter to add, Backspace to remove last)
- Works with react-hook-form

Add example to styleguide.

### 7. Frontend: Add useUserTags hook

**Location:** `frontend/apps/webv2/app/immersion/api.ts`

- API client for GET `/users/{userId}/tags`
- React Query hook with debounced prefix filtering

### 8. Frontend: Update log form to use TagInput

**Location:** `frontend/apps/webv2/app/immersion/NewLogForm/`

In `domain.tsx`:
- Change schema from `z.array(Tag).max(3)` to `z.array(z.string()).max(10)`
- Remove `Tag` type and `filterTags` function

In `Form.tsx`:
- Replace `AutocompleteMultiInput` with new `TagInput`
- Wire up `useUserTags` for autocomplete suggestions
- Remove activity-based tag filtering

### 9. Backend: Remove tags from LogConfigurationOptions

**Location:** `services/immersion-api/http/rest/server_loggetconfigurations.go`

Remove tags from this endpoint (frontend uses new endpoint instead).

### 10. Testing

- Backend: Test tag validation edge cases
- Backend: Test tag queries and frequency ordering
- Frontend: Test TagInput keyboard navigation, mobile, accessibility
- E2E: Submit log with custom tags, verify autocomplete and display

---

## Files to modify

**Backend:**
- `services/immersion-api/storage/postgres/migrations/` (new migration)
- `services/immersion-api/storage/postgres/queries/logs.sql`
- `services/immersion-api/storage/postgres/queries/tags.sql`
- `services/immersion-api/domain/command/createlog.go`
- `services/immersion-api/storage/postgres/repository/repo_createlog.go`
- `services/immersion-api/http/rest/server_getusertags.go` (new)
- `services/immersion-api/http/rest/server_loggetconfigurations.go`
- `services/immersion-api/storage/postgres/models.go`

**Frontend:**
- `frontend/packages/ui/components/Form/TagInput.tsx` (new)
- `frontend/apps/webv2/app/immersion/api.ts`
- `frontend/apps/webv2/app/immersion/NewLogForm/domain.tsx`
- `frontend/apps/webv2/app/immersion/NewLogForm/Form.tsx`
- `frontend/apps/styleguide/pages/forms.tsx`
