# Account Deletion Guide

This document provides the SQL queries and API commands needed to completely delete a user account from Tadoku.

## Overview

Deleting an account involves:
1. Deleting all user data from the immersion database (PostgreSQL)
2. Deleting the identity from Kratos (identity management)

## Prerequisites

- Database access to the immersion database
- Access to Kratos Admin API
- The user's account ID (UUID)

## Step 1: Delete Data from Immersion Database

Run the following SQL transaction in the immersion database. Replace `<ACCOUNT_ID>` with the actual user UUID at the top.

```sql
-- Define the account ID to delete here
SET session.account_to_delete = '<ACCOUNT_ID>';

BEGIN;

-- Set the schema search path
SET LOCAL search_path TO data;

DO $$
DECLARE
    v_account_id UUID := current_setting('session.account_to_delete')::UUID;
    finished_contest_logs_count INTEGER;
BEGIN
    -- Step 0: Check if user has logs attached to finished contests
    -- If this query returns any rows, ABORT the deletion
    SELECT COUNT(*)
    INTO finished_contest_logs_count
    FROM contest_logs cl
    JOIN logs l ON cl.log_id = l.id
    JOIN contests c ON cl.contest_id = c.id
    WHERE l.user_id = v_account_id
      AND c.contest_end < NOW()
      AND c.deleted_at IS NULL;

    IF finished_contest_logs_count > 0 THEN
        RAISE EXCEPTION 'Cannot delete account: User has % logs attached to finished contests. Use account anonymization instead.', finished_contest_logs_count;
    END IF;

    -- Step 1: Delete contest_logs for user's logs
    DELETE FROM contest_logs
    WHERE log_id IN (SELECT id FROM logs WHERE user_id = v_account_id);

    -- Step 2: Delete user's logs (hard delete)
    DELETE FROM logs WHERE user_id = v_account_id;

    -- Step 3: Delete user's contest registrations
    DELETE FROM contest_registrations WHERE user_id = v_account_id;

    -- Step 4: Anonymize contests owned by the user
    -- (preserves contest data for other participants)
    UPDATE contests
    SET owner_user_id = '00000000-0000-0000-0000-000000000000',
        owner_user_display_name = '[deleted user]'
    WHERE owner_user_id = v_account_id;

    -- Step 5: Delete user record
    DELETE FROM users WHERE id = v_account_id;
END $$;

COMMIT;

-- Clean up the session variable
RESET session.account_to_delete;
```

### Alternative: Delete Contests Instead of Anonymizing

If you prefer to completely delete contests owned by the user (note: this affects other participants), replace Step 4 with:

```sql
-- Step 4a: Delete contest_logs for user's contests
DELETE FROM contest_logs
WHERE contest_id IN (SELECT id FROM contests WHERE owner_user_id = '<ACCOUNT_ID>');

-- Step 4b: Delete registrations for user's contests
DELETE FROM contest_registrations
WHERE contest_id IN (SELECT id FROM contests WHERE owner_user_id = '<ACCOUNT_ID>');

-- Step 4c: Delete contests
DELETE FROM contests WHERE owner_user_id = '<ACCOUNT_ID>';
```

## Step 2: Delete Identity from Kratos

Use the Kratos Admin API to delete the identity. Replace `<KRATOS_ADMIN_URL>` with your Kratos admin endpoint and `<ACCOUNT_ID>` with the user UUID.

```bash
curl -X DELETE \
  '<KRATOS_ADMIN_URL>/admin/identities/<ACCOUNT_ID>' \
  -H 'Accept: application/json'
```

### Example

```bash
curl -X DELETE \
  'http://localhost:4434/admin/identities/a1b2c3d4-e5f6-7890-abcd-ef1234567890' \
  -H 'Accept: application/json'
```

## Verification

After deletion, verify the account is gone:

```sql
-- Should return no rows
SELECT * FROM users WHERE id = '<ACCOUNT_ID>';
```

```bash
# Should return 404
curl -X GET \
  '<KRATOS_ADMIN_URL>/admin/identities/<ACCOUNT_ID>' \
  -H 'Accept: application/json'
```

## Notes

- **Moderation audit logs** are NOT deleted or anonymized by default. If needed, add:
  ```sql
  DELETE FROM moderation_audit_log WHERE user_id = '<ACCOUNT_ID>';
  ```

- **Contest anonymization** preserves contests for other participants while removing the deleted user's identity

- **Transaction safety**: The SQL runs in a transaction, so if any step fails, all changes are rolled back

- **Kratos cleanup**: If the Kratos deletion fails but database deletion succeeded, you can retry the curl command safely

## Account Anonymization (For Users with Finished Contest Logs)

**⚠️ TODO: Not Yet Implemented**

If a user has logs attached to finished contests, complete deletion is not possible because it would corrupt historical contest data. Instead, the account must be **anonymized**.

### When to Use Anonymization

The deletion query will abort with an error if the user has logs in finished contests. In this case, you must use account anonymization instead.

### What Anonymization Should Do

The following functionality needs to be developed:

1. **Anonymize user record**:
   - Set `display_name` to `[deleted user <random-id>]` to prevent display name collisions
   - Remove any PII fields if they exist

2. **Anonymize logs**:
   - Keep logs intact (required for contest history)
   - Consider adding a `anonymized` flag or field

3. **Anonymize contest ownership** (same as deletion):
   - Set `owner_user_id` to null UUID
   - Set `owner_user_display_name` to `[deleted user]`

4. **Delete contest registrations**:
   - Remove from future/ongoing contests
   - Keep registrations for finished contests (or anonymize if needed)

5. **Delete Kratos identity**:
   - Prevent login while preserving data integrity

6. **Mark account as deleted**:
   - Consider adding a `deleted_at` timestamp to the `users` table
   - Ensure anonymized users can't be reactivated

### Implementation Status

This query has not been implemented yet. When a user with finished contest logs needs to be deleted, you will need to:

1. Develop the anonymization query following the guidelines above
2. Test thoroughly to ensure contest data integrity is preserved
3. Update this documentation with the final query

## Troubleshooting

### "Cannot delete account: User has logs attached to finished contests"

This error means the user has participated in contests that have already ended. Complete deletion would corrupt historical contest data.

**Solution**: Use the account anonymization approach instead (see section above - currently TODO).

### Database deletion fails

Check for foreign key constraints or active connections. Ensure no other processes are modifying the user's data during deletion.

### Kratos deletion fails with 404

The identity may already be deleted or never existed. This is safe to ignore if the database deletion succeeded.

### Rollback needed

If you need to rollback the database transaction, run:
```sql
ROLLBACK;
```
