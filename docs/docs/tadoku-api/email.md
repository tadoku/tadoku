---
title: Retained email
description: Transactional template rendering, SMTP delivery and holds, administrator reads and audit records, and tenant-scoped email retention.
---

# Retained email

Read this when a business operation sends email, when inspecting retained email,
or when configuring delivery. The feature is independent of account workflows.
Its writer, reader and delivery variants live in
`services/tadoku-api/features/email/`; the templates are plain text.

## Write inside the business transaction

Pass a `domain/mail.Message` to `email.Writer.Write` inside the caller's existing
PostgreSQL transaction. Persist every returned `WriteResult.Jobs` through
`jobqueue.Service.Enqueue` before committing that same transaction. Do not call
the writer outside a transaction or enqueue after commit. Failure rolls back
the business write, email and queued work together.

| Template | Required string parameters |
| --- | --- |
| `account.recovery` | `recovery_url` |
| `account.email_changed` | `new_email`, `changed_at` (RFC3339) |

The writer rejects unknown templates, missing, empty or unexpected parameters,
parameters over 4096 bytes, invalid addresses and recovery links outside the
Tadoku app host allowlist. Addresses are bare mailboxes, without display names
or header line breaks. Recovery links require HTTPS and exclude credentials
and explicit ports. The allowlist covers the main, account, admin and authv2
hosts, preview hosts and development branch equivalents.

A successful write inserts a `pending` email with its rendered subject/body and
parameter map. The job is `email.send.v1` with only `{"email_id":"<uuid>"}`.
Its tenant is the queue row's tenant; payloads never carry tenant, recipient,
content or credentials. See [Jobs and worker](./jobs.md#email-message-contract).

Construct the service with `NewService(repository, sender)`. A nil sender
disables SMTP and holds every pending email.

## Delivery and failures

The worker registers email delivery with concurrency 2, a 20-second handler
bound and five attempts. It holds a pending row unless both the deployment
has enabled sending and the persisted tenant is `tadoku/prod`. A missing ID in
the job's tenant is harmless; a guessed ID cannot reach another tenant's row.
Terminal rows are not sent again.

Enabled production delivery uses authenticated SMTP over STARTTLS, verifies
the server certificate and records `sent`, the sending time and the provider's
message ID returned by SMTP. The sender supplies an RFC Message-ID derived from
the email ID, but [SES replaces Message-ID and Date headers](https://docs.aws.amazon.com/ses/latest/dg/header-fields.html).
Use the recorded provider message ID to investigate a delivered message.
A failed SMTP attempt leaves the email pending until the fifth attempt, which
records `failed` and its error. Worker maintenance also reconciles pending rows
whose corresponding delivery jobs have exhausted their attempts, including
abandoned final leases. Reconciliation runs at startup and with the worker's
15-second maintenance pass.

SMTP cannot make the external send and PostgreSQL update atomic. A crash after
SMTP accepts a message and before `sent` is persisted can deliver twice. The
submitted Message-ID does not provide a deduplication guarantee.
See [Execution guarantees](./jobs.md#execution-and-failure-guarantees).

## Read retained messages

`GET /emails` and `GET /emails/{id}` require an administrator of the verified
request tenant. The gateway must expose these API routes; an environment without
its email rule returns 404 before reaching the API. Administrator cookie requests
require the gateway's `cookie_session` authentication and `id_token` mutation.
The list accepts exact recipient, template and status filters,
`since` (inclusive), `limit` (1–100; default 100) and `cursor`. Results are newest
first, with ID as a deterministic tie-breaker. Continue with the opaque
`next_cursor`; cursors belong to one tenant and never change request scope.
Both endpoints return rendered content and params, so treat the response as
sensitive even in development.

Every returned `tadoku/prod` row writes one `read_email` moderation audit entry,
containing the actor and email ID, in the read transaction. A failed audit write
fails the response without returning email content. Empty lists and denied or
missing reads disclose no email and create no read entry. Other tenants do not
write email-read audits.

Non-production `email:read` access uses the dedicated token route and a
tenant-bound verifier. An absent verifier fails closed. Canonical development
token reads record the token UUID as actor with `access_token_id` metadata in the
same fail-closed audit transaction; production rejects those tokens. Gateway bearer pass-through preserves that verifier boundary alongside
administrator cookie authentication; anonymous access never substitutes for token
verification. See [access-token commands](./tenant-lifecycle.md#email-access-tokens).

## Retention and operations

The worker deletes production rows strictly older than three days, and other
tenants' rows strictly older than fourteen days. Exact-boundary rows remain.
It runs retention at startup and daily. A branch worker maintains only its own
tenant; the base enumerates registered tenants and runs each statement within
that tenant's ordinary row-level security scope. Deleting a tenant cascades to
its email rows. Email retention and successful-job retention are independent.

See [Worker configuration](./worker-configuration.md#email-delivery) for settings.
Local E2Es use disposable databases, real worker execution, real authorization
providers and a loopback SMTP server with TLS and synthetic credentials. They
do not contact SES. Run them with the disposable provider setup in
[Testing](./testing.md) to repeat the writer → worker → administrator-read proof
without shared provider writes. Live branch verification uses an already approved
session or branch-owned account fixture and confirms the selected API backend.
Creating a shared-provider session requires explicit maintainer authorization.
If an account flow needs a temporary administrator-only test entrypoint, restrict
it to the branch, remove it afterward and retain its exact patch and sanitized
responses outside source control. Never add a permanent public send endpoint or
change shared provider identities to manufacture verification.
