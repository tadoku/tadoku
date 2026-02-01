# notification-api Design Document

## Overview

A queue-based email notification service using AWS SES for delivery. Other services send email requests via HTTP, which are queued and processed asynchronously.

## Architecture

```
┌─────────────────┐     HTTP      ┌──────────────────────────────────┐
│  immersion-api  │──────────────▶│         notification-api         │
│  content-api    │               │                                  │
└─────────────────┘               │  ┌────────────┐   ┌───────────┐  │
                                  │  │ HTTP API   │──▶│  Enqueue  │  │
                                  │  └────────────┘   └─────┬─────┘  │
                                  │                         │        │
                                  │                   ┌─────▼─────┐  │
                                  │                   │   Queue   │  │
                                  │                   │ (Postgres)│  │
                                  │                   └─────┬─────┘  │
                                  │                         │        │
                                  │  ┌────────────┐   ┌─────▼─────┐  │
                                  │  │ Worker     │◀──│  Dequeue  │  │
                                  │  │ (goroutine)│   └───────────┘  │
                                  │  └─────┬──────┘                  │
                                  └────────┼─────────────────────────┘
                                           │
                                  ┌────────▼────────┐
                                  │     AWS SES     │
                                  └─────────────────┘
```

## Components

### 1. HTTP API

Receives email requests from internal services. Validates input and enqueues for processing.

**Endpoint:** `POST /v1/emails`

```json
{
  "to": "user@example.com",
  "subject": "Contest Starting Soon",
  "body_text": "The Summer Reading Contest starts in 1 hour!",
  "body_html": "<h1>Contest Starting Soon</h1><p>The Summer Reading Contest starts in 1 hour!</p>",
  "from": "contests@tadoku.app",
  "reply_to": "support@tadoku.app",
  "priority": "normal",
  "scheduled_at": "2024-01-15T10:00:00Z",
  "idempotency_key": "contest-123-start-reminder-user-456"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `to` | Yes | Recipient email address |
| `subject` | Yes | Email subject line |
| `body_text` | Yes | Plain text body (fallback) |
| `body_html` | No | HTML body |
| `from` | No | Sender address (default: configured default) |
| `reply_to` | No | Reply-to address |
| `priority` | No | `high`, `normal`, `low` (default: `normal`) |
| `scheduled_at` | No | Send at specific time (default: immediate) |
| `idempotency_key` | No | Prevent duplicate sends |

**Response:** `202 Accepted`
```json
{
  "id": "email_abc123",
  "status": "queued"
}
```

### 2. Queue (PostgreSQL-based)

Using PostgreSQL for the queue keeps infrastructure simple (no new dependencies). The queue supports:

- Priority ordering (high → normal → low)
- Scheduled delivery
- Retry with exponential backoff
- Idempotency

**Table: `email_queue`**

```sql
CREATE TABLE email_queue (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(255) UNIQUE,

    -- Email content
    to_email        VARCHAR(255) NOT NULL,
    from_email      VARCHAR(255) NOT NULL,
    reply_to        VARCHAR(255),
    subject         VARCHAR(998) NOT NULL,  -- RFC 2822 limit
    body_text       TEXT NOT NULL,
    body_html       TEXT,

    -- Queue management
    priority        SMALLINT NOT NULL DEFAULT 1,  -- 0=high, 1=normal, 2=low
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    scheduled_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Processing tracking
    attempts        SMALLINT NOT NULL DEFAULT 0,
    max_attempts    SMALLINT NOT NULL DEFAULT 3,
    last_error      TEXT,
    locked_until    TIMESTAMPTZ,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at         TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT valid_status CHECK (status IN ('pending', 'processing', 'sent', 'failed', 'cancelled'))
);

-- Index for efficient queue polling
CREATE INDEX idx_email_queue_pending ON email_queue (priority, scheduled_at)
    WHERE status = 'pending' AND (locked_until IS NULL OR locked_until < NOW());
```

**Status Flow:**
```
pending → processing → sent
                    ↘ failed (after max retries)
```

### 3. Worker

A background goroutine that:

1. Polls the queue for pending emails
2. Locks rows to prevent duplicate processing
3. Sends via AWS SES
4. Updates status (sent/failed)
5. Implements exponential backoff for retries

**Polling Query:**
```sql
UPDATE email_queue
SET status = 'processing',
    locked_until = NOW() + INTERVAL '5 minutes',
    attempts = attempts + 1,
    updated_at = NOW()
WHERE id = (
    SELECT id FROM email_queue
    WHERE status = 'pending'
      AND scheduled_at <= NOW()
      AND (locked_until IS NULL OR locked_until < NOW())
      AND attempts < max_attempts
    ORDER BY priority, scheduled_at
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;
```

**Retry Backoff:**
- Attempt 1: Immediate
- Attempt 2: 1 minute delay
- Attempt 3: 5 minutes delay
- After 3 failures: Mark as `failed`

### 4. AWS SES Client

Wrapper around AWS SDK for sending emails.

```go
type SESClient interface {
    SendEmail(ctx context.Context, email Email) error
}
```

## Configuration

```yaml
# Environment variables
NOTIFICATION_API_PORT: "8080"
NOTIFICATION_API_DATABASE_URL: "postgres://..."

# AWS SES
AWS_REGION: "us-east-1"
AWS_ACCESS_KEY_ID: "..."
AWS_SECRET_ACCESS_KEY: "..."

# Email defaults
NOTIFICATION_DEFAULT_FROM: "noreply@tadoku.app"
NOTIFICATION_DEFAULT_REPLY_TO: "support@tadoku.app"

# Worker settings
NOTIFICATION_WORKER_POLL_INTERVAL: "5s"
NOTIFICATION_WORKER_BATCH_SIZE: "10"
```

## Directory Structure

```
services/notification-api/
├── main.go
├── domain/
│   ├── models.go           # Email, EmailRequest, etc.
│   ├── interfaces.go       # Repository, SESClient interfaces
│   ├── errors.go
│   ├── emailenqueue.go     # Enqueue service
│   └── emailprocess.go     # Worker/processor service
├── http/rest/
│   ├── server.go
│   ├── server_emailcreate.go
│   └── openapi/
│       ├── spec.yaml
│       └── generate.go
├── storage/postgres/
│   ├── migrations/
│   │   └── 001_create_email_queue.sql
│   ├── queries.sql
│   ├── repository/
│   │   └── email.go
│   └── generate.go
├── client/
│   └── ses/
│       └── client.go
└── worker/
    └── worker.go
```

## Future Considerations (Deferred)

### Unsubscribe Handling

When implementing unsubscribe:

1. **Signed Unsubscribe Links**
   - Generate HMAC-signed URLs: `/unsubscribe?token=<signed_token>`
   - Token contains: user_id, email_type, expiry
   - No login required to unsubscribe

2. **Email Preferences**
   - Stored in auth app (as specified)
   - notification-api checks preferences before sending
   - API endpoint for auth app to query/update preferences

3. **List-Unsubscribe Header**
   - Add RFC 8058 compliant headers for one-click unsubscribe

### Batch Sending

For sending to multiple recipients (e.g., contest announcements):

```json
POST /v1/emails/batch
{
  "recipients": ["user1@example.com", "user2@example.com"],
  "subject": "...",
  "body_text": "..."
}
```

### Scheduled/Recurring Emails

For digest emails:
- Cron-like scheduling
- Template support with data fetching

### Delivery Tracking

If needed later:
- Webhook integration with SES for bounce/complaint handling
- Store delivery events in separate table

## Security Considerations

1. **Internal Only**: This API should only be accessible from internal services (not public)
2. **Rate Limiting**: Implement per-service rate limits
3. **Input Validation**: Validate email addresses, sanitize HTML
4. **Idempotency**: Prevent duplicate sends with idempotency keys
5. **Secrets**: AWS credentials via environment variables or IAM roles

## Implementation Phases

### Phase 1: Core Infrastructure ✨ (Current Focus)
- [ ] Project scaffolding
- [ ] Database schema and migrations
- [ ] Email queue repository
- [ ] HTTP API for enqueueing
- [ ] Worker for processing queue
- [ ] AWS SES integration
- [ ] Basic error handling and retries

### Phase 2: Reliability
- [ ] Idempotency key support
- [ ] Improved retry logic with backoff
- [ ] Dead letter queue for failed emails
- [ ] Health check endpoint
- [ ] Metrics/logging

### Phase 3: Unsubscribe & Preferences
- [ ] Signed unsubscribe links
- [ ] Unsubscribe endpoint (no auth required)
- [ ] Integration with auth app for preferences
- [ ] List-Unsubscribe headers

### Phase 4: Advanced Features
- [ ] Batch sending
- [ ] Scheduled/recurring emails
- [ ] Delivery tracking (bounces, complaints)
