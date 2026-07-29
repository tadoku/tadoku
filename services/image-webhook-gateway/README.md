# Image Webhook Gateway

This service serializes registry-triggered Argo CD Image Updater
reconciliations for one `ImageUpdater` resource.

## Why it exists

Argo CD Image Updater acknowledges a valid webhook before the reconciliation
and Git write finish. Several registry events can therefore receive successful
HTTP responses and continue processing concurrently. When those reconciliations
write to the same Git branch, their pushes can race and fail with a
non-fast-forward error.

Limiting concurrent HTTP connections does not solve this problem because the
webhook connection closes before the asynchronous work completes. This gateway
holds a FIFO slot until the corresponding Image Updater status reports that the
reconciliation has finished.

## How it works

For each incoming request, the gateway:

1. verifies the GitHub HMAC signature;
2. accepts only published container-package events for the `prod` tag;
3. stores the original request body in one bounded, in-memory FIFO;
4. forwards that unchanged body to the internal Image Updater webhook;
5. waits for the configured `ImageUpdater` resource's `lastCheckedAt` value to
   advance and its `Reconciling` condition to become `False`;
6. processes the next queued event.

The forwarded request is signed again over the original bytes and retains the
GitHub event and delivery headers.

## Scope

The gateway serializes webhook-triggered reconciliations handled by its single
worker. It does not schedule image checks. Image Updater retains responsibility
for its native periodic recovery reconciliation, metrics, and alerts.

The native periodic reconciliation is independent of the gateway FIFO. The
gateway waits if it observes an active reconciliation, but it cannot provide a
shared lock to Image Updater's internal scheduler without changing Image
Updater itself.

## Configuration

Configuration uses `envconfig` with the `GATEWAY` prefix. Every value is
required; this repository intentionally provides no environment-specific
defaults.

| Variable | Purpose |
| --- | --- |
| `GATEWAY_ADDRESS` | HTTP listen address |
| `GATEWAY_GHCR_WEBHOOK_SECRET` | Shared secret used to verify and re-sign webhook bodies |
| `GATEWAY_IMAGE_UPDATER_WEBHOOK_URL` | Internal Image Updater webhook endpoint |
| `GATEWAY_IMAGE_UPDATER_NAMESPACE` | Namespace containing the observed `ImageUpdater` |
| `GATEWAY_IMAGE_UPDATER_NAME` | Name of the observed `ImageUpdater` |
| `GATEWAY_QUEUE_SIZE` | Maximum number of webhook events waiting in memory |

Deployment-specific values belong in the private GitOps repository.

## HTTP endpoints

| Path | Purpose |
| --- | --- |
| `/webhook` | Receives GitHub package webhooks |
| `/healthz` | Liveness probe |
| `/readyz` | Readiness probe; becomes unavailable during graceful shutdown |

Accepted events return `202` after entering the queue. If the queue is full,
the gateway returns `503` so the sender can retry. Valid non-`prod` events
return `202` without entering the queue.

## Shutdown and recovery

On `SIGTERM`, the gateway stops accepting requests, closes its HTTP server, and
drains the current in-memory queue before exiting. Deployment configuration must
provide enough termination grace for the drain to complete and must avoid
overlapping gateway replicas.

An unexpected process or node failure can lose queued events. Image Updater's
native periodic reconciliation is the recovery path for that case.

## Development

```sh
bazel test //services/image-webhook-gateway:all
bazel build //services/image-webhook-gateway:image
bazel run //:gazelle -- -mode=diff
```
