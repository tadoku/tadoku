---
title: Identity contracts
description: Identity, email and tenant access-token service contracts, their stub routes and authentication boundaries, and the temporary v1 reader.
---

# Identity contracts

Read this before implementing account, recovery, email or tenant access-token
operations, or moving a caller to the identity entry point.

The identity command/session, email and token services are skeletons: new operations return
`errx.NotImplemented`, mapped to `501` with `{"message":"not implemented"}`.
No account, cookie, recovery token, email or access token is created or changed.
Profile and moderation callers use `identity.Reader`, constructed by
`identity.NewReader`. It selects tenant-local reads or the v1 adapter from the
tenant flag; callers never select the provider.

## Routes

| Group | Operations |
| --- | --- |
| JWT and ban gate | `POST /identity/registration`, `POST /identity/login`, `POST /identity/logout`, `GET /identity/session`, `POST /identity/recovery`, `GET /identity/recovery/{token}`, `PATCH /identity/settings/profile`, `PUT /identity/settings/password` |
| JWT and ban gate | `GET /emails`, `GET /emails/{id}` |
| Oathkeeper callback bearer | `GET /identity/internal/v1/session` |
| Email-only token registrar | The same two email GET routes, selected by a `Bearer tdk_at_…` header |

Public route paths are relative to `/api/internal`. They require the existing
JWT even for guests. Missing or invalid JWT credentials are rejected before the
stub. Callback requests require `API_OATHKEEPER_AUTHZ_TOKEN`, without creating a
user identity. Email tokens enter `Application.AuthenticateEmailReader`, which
returns `501` before decoding or authorizing anything. No token currently
succeeds. A token-shaped bearer on every other public route remains an invalid
JWT. Gateway token forwarding is not enabled by these skeletons.

The canonical fragments in `services/tadoku-api/spec/openapi/` define success,
error, cookie and redirect shapes as well as email filters and pagination.
`GET /identity/session` carries `account_app` (`v1` or `v2`) and a nullable
account. The callback preserves the Kratos whoami user shape; its guest variant
has `identity.id` equal to `guest`.

## Service boundaries

| Owner | Contract |
| --- | --- |
| `services/tadoku-api/features/identity/` | `Reader` owns account lookup, existence and search; `Accounts` adds registration, login, logout, session resolution, recovery, settings and deletion. `ReadService` implements reads; `Service` implements the command/session stubs. |
| `services/tadoku-api/features/email/` | `Writer.Write(ctx, mail.Message)` renders and writes within the caller's transaction and returns every required job. `Reader` owns filtered listing and lookup. `Service.Send` and `Retain` are worker operation stubs. |
| `services/tadoku-api/features/authz/tokens/` | `Verifier.Verify` accepts the raw token and deployment scope; `Manager` owns create, list and revoke. Results expose a parsed tenant, ID and scopes, never stored hashes. |
| `services/tadoku-api/domain/mail/` | `Message` is the application-composed template request shared by identity and email. Templates are `account.recovery` (`recovery_url`) and `account.email_changed` (`new_email`, `changed_at`). |
| `services/tadoku-api/domain/jobs/` | `SendEmailV1` carries only `email_id`; see [Jobs](./jobs.md#email-message-contract). |

Transport inputs remain untrusted command values. Feature operations own their
parsing and validation. `Email`, `DisplayName` and `Password` have private
representations and explicit parsing constructors; those constructors also return
not implemented until validation is implemented. `Account` returns the typed
email and display name. Password formatting is redacted. Guest/authenticated sessions and pending/sent/held/failed
email delivery results are explicit variants with their required payloads.
Password hashes and session/recovery token hashes never cross these interfaces.

`Accounts.Logout`, `Session`, `CheckSession`, `UpdateProfile` and
`UpdatePassword` receive `cookieHeader` of type `CookieHeader`, the complete
incoming HTTP `Cookie` header, including cookie names. It is neither a single
cookie value nor an Authorization bearer. The identity service resolves the tenant from `ctx` and
selects the implementation using the tenant's `identity-v2` decision. The v1
adapter selects `ory_kratos_session`; v2 selects `tadoku_session` for
`tadoku/prod`, or `tadoku_session_` plus the tenant key with `/` replaced by `_`
for other tenants. Callers must not select the cookie or switch implementations.
`ConsumeRecovery.recoveryToken` has type `RecoveryToken`, the raw token from the
URL path, without a cookie name or bearer prefix. `UpdatePassword.password` is the
untrusted new plaintext password; the feature validates it before use. The
cookie header, recovery token and password must never be logged. `CookieHeader`
and `RecoveryToken` distinguish raw credentials; they do not imply validation.

Identity recovery and profile updates return mail requests; application code
coordinates the identity operation, each email write and every returned queue
job inside one business transaction. Features do not import sibling features.
The email writer must reject unknown templates and missing or unexpected
parameters before inserting. These behaviors are not implemented by the stubs.

The `identity-v2` flag is a tenant decision with safe default `false`.
`featureflags.Service.IdentityV2(ctx)` evaluates the context tenant, including
for guests. The session contract will expose account-app selection; adding this
flag does not change the existing public decisions response or golden fixtures.

## V1 adapter

`identity.NewLegacyReader` constructs the unchanged read-only Kratos adapter
behind `LegacyReader`. Its raw implementation and provider tests live in
`services/tadoku-api/features/identity/kratos/`. Only identity ownership and the owned Kratos fixture consume this temporary
interface. Profile and authz receive provider-free `Reader` values. The unused
guarded Kratos writer and the profile user cache have been removed.

V1 reads retain the user-schema filter, trait mapping, provider order and fuzzy
search ranking. Listing reads the provider's complete cursor sequence per call,
deduplicates IDs and rejects repeated continuation tokens; no stale snapshot is
served after an outage. V2 reads join active accounts to undeleted users under
the transaction-local tenant. Each reader page uses one SQL statement, orders by
user UUID and searches display name or email with a literal case-insensitive
substring. Cursors bind the tenant, search and last UUID. No password or token
hash is selected.

The administrator HTTP list retains numbered pages and `total_size` for
compatibility. Its adapter walks reader pages to count matches and select the
requested numbered page, then resolves roles only for that page. This can read
multiple keyset pages for a large result set; the public response shape is
unchanged.
