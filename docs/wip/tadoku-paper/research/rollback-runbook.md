# Tadoku Paper deployment rollback runbook

Status: Phase 0 operational contract

## Safety model

Paper delivery changes only static/frontend artifacts. It has no database, schema, API, or Ory migration. A failed Paper or per-application visual cutover rolls back by deploying the previously recorded immutable image digest and reverting source normally. Never recover by loading legacy and Paper stylesheets together.

Production is self-healing GitOps. `kubectl set image`, `kubectl rollout undo`, and hand-edited live resources are temporary and will be reverted by Argo CD. Every durable rollback is a reviewed change to `antonve/tadoku-argocd`.

The Image Updater follows the mutable `prod` tag. Pinning an older digest while its Application remains in the updater list is unstable: the next webhook or hourly reconciliation can reapply the bad `prod` digest. Every emergency rollback PR must both pause that application's updater mapping and pin the previous digest.

## Required release record

Create a gate record before every production cutover:

| Field | Required value |
| --- | --- |
| Target | Paper, admin, auth, webv2, or final UI |
| Tadoku source commit | full 40-character SHA |
| Incoming image | repository plus full `sha256:` digest |
| Outgoing image | repository plus full `sha256:` digest |
| GitOps pre-change revision | full SHA |
| Argo CD Application | exact name |
| Namespace / Deployment | exact names |
| Public smoke routes | target-specific list |
| Updater mapping | exact `namePattern` and image alias |
| Rollback PR | prepared branch/PR or exact two-file patch plan |
| Operator and UTC time | human/agent identity and timestamp |

Do not infer the outgoing digest from `prod`, `latest`, a Deployment annotation, or a source commit. Read it from the running pod and reconcile it with the GitOps Kustomize digest.

Example read-only capture shape:

```sh
kubectl --context lke20948-ctx -n tdk-prod-paper-styleguide \
  get pod -l app=tadoku-paper-styleguide \
  -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.containerStatuses[0].imageID}{"\n"}{end}'

kubectl --context lke20948-ctx -n argocd \
  get application tadoku-paper-styleguide \
  -o jsonpath='{.status.sync.status}{" "}{.status.health.status}{" "}{.status.sync.revision}{"\n"}'
```

## Rollback decision

Roll back immediately when any of these occurs after deployment:

- the Application cannot become `Synced Healthy` within the recorded rollout budget;
- readiness or liveness fails, a pod restarts, or no ready replica remains;
- the public certificate or host route is wrong;
- `/`, a canonical deep link, or required assets fail;
- missing hashed assets return HTML instead of `404`;
- keyboard/navigation/search behavior blocks catalogue use;
- error telemetry materially regresses;
- CPU throttling, OOM, or measured usage violates the approved resource envelope.

Minor documentation defects that do not violate the gate can use the normal forward-fix path.

## Standard Paper digest rollback

Use this for an established `paper.tadoku.app` deployment.

1. Stop further Tadoku merges that publish `frontend-paper-styleguide:prod`.
2. Open an emergency GitOps PR that:
   - removes or disables only the `tadoku-paper-styleguide` entry in `argocd/image-updater/image-updater.yaml`; and
   - sets `services/paper-styleguide/kustomization.yaml` to the recorded previous image name and digest.
3. Update `scripts/refactor/validate-image-updater.sh` in the same PR if its exact expected-app policy requires the paused entry to be absent. Do not weaken unrelated updater checks.
4. Run `make validate`; merge the PR; record its SHA.
5. Observe Argo CD rather than mutating the live Deployment:

```sh
kubectl --context lke20948-ctx -n argocd \
  get application tadoku-paper-styleguide -w

kubectl --context lke20948-ctx -n tdk-prod-paper-styleguide \
  rollout status deployment/tadoku-paper-styleguide --timeout=5m
```

6. Confirm every running pod reports the previous digest, then rerun certificate, `/`, deep-link, health, asset, and missing-asset smoke checks.
7. Revert or fix the Tadoku source through a normal PR. Its successful main build produces a new immutable digest and advances `prod`.
8. Verify that new digest manually at `paper.tadoku.app` before re-enabling the Paper updater mapping in a separate GitOps PR.

Do not re-enable the updater while `prod` still resolves to the failed digest.

## First Paper deployment abort

The first deployment has no previous Paper image. If it fails before being declared live:

1. Pause/remove the Paper Image Updater mapping.
2. Revert the GitOps PR that introduced the Paper Application and service root, preserving unrelated commits.
3. Let Argo CD prune only the Paper resources. The Namespace carries `Delete=confirm`, so namespace deletion remains explicitly guarded.
4. Confirm there is no Paper Ingress or certificate and that other production Applications remain healthy.
5. Leave the existing DNS A record unchanged. It existed before Paper and will return Kong's default 404 until a corrected deployment is ready.

No legacy `ui.tadoku.app` traffic is involved.

## Final `ui.tadoku.app` cutover rollback

The final cutover keeps the existing `tdk-prod-styleguide/tadoku-ui` Service, Ingress, certificate, and DNS, and changes only the selected Deployment image/configuration. This makes rollback one workload change.

Before cutover, record both:

- legacy image `ghcr.io/tadoku/tadoku/frontend-styleguide@sha256:...` from the running pod; and
- verified Paper image `ghcr.io/tadoku/tadoku/frontend-paper-styleguide@sha256:...` already serving `paper.tadoku.app`.

If the UI smoke gate fails:

1. Keep the existing styleguide updater mapping paused.
2. Open a GitOps rollback PR restoring the legacy image name/digest, port 3000, legacy probe paths, legacy resource envelope, and any legacy security/runtime fields in `services/styleguide/deployment.yaml` and `kustomization.yaml`.
3. Run `make validate`, merge, and observe `tadoku-styleguide` become `Synced Healthy`.
4. Confirm `ui.tadoku.app` serves the legacy digest and its known routes.
5. Confirm `paper.tadoku.app` still serves the independent Paper Deployment; it is not part of this rollback.
6. Leave the legacy source and workflow intact until a later cutover passes its full smoke window.

Do not alter DNS or the `ui.tadoku.app` TLS Secret during either cutover or rollback. Do not create a competing `ui.tadoku.app` Ingress in the Paper namespace.

## Application migration rollback (future phases)

Admin, auth, and webv2 retain their host, Service, APIs, cookies, and database contracts. For each application:

1. pause that application's Image Updater mapping;
2. pin its recorded previous digest in its existing GitOps Kustomization;
3. validate and merge the GitOps rollback PR;
4. verify the running pod digest and application-specific production smoke list;
5. revert/fix the Tadoku migration PR normally;
6. publish and verify a corrected `prod` digest before re-enabling updates.

Previously deployed images contain their own legacy dependencies, so rollback does not require restoring deleted source to the currently checked-out repository. Nevertheless, Phase 8 legacy source deletion must not occur until every application smoke window has closed.

## Verification checklist

- [ ] Image Updater cannot overwrite the rollback pin.
- [ ] GitOps `main` contains the intended previous digest.
- [ ] Argo CD reports `Synced Healthy` at the rollback Git revision.
- [ ] Every ready pod's `imageID` matches the recorded digest.
- [ ] Rollout completed with an available replica and no new restarts/OOMs.
- [ ] Public TLS, root, canonical deep links, assets, and health checks pass.
- [ ] Target-specific keyboard/form/role smoke checks pass.
- [ ] Error/resource telemetry returned to baseline.
- [ ] The incident record includes incoming/outgoing digests and both source revisions.
- [ ] Updater remains paused until a separately verified fixed `prod` digest exists.

## Known operational hazards

| Hazard | Control |
| --- | --- |
| Mutable `prod` immediately defeats an old digest pin | Pause the exact updater entry in the rollback PR. |
| Direct `kubectl` rollback appears successful briefly | Treat Argo CD as authoritative; all durable changes go through Git. |
| Kustomization digest and Deployment literal differ | Inspect rendered/live image; Kustomize override wins. |
| Two Ingresses claim `ui.tadoku.app` | Keep final cutover inside the existing styleguide workload boundary. |
| Missing Vite favicon makes old liveness probe fail | Restore the complete runtime-specific probe set with the image. |
| Re-enabling updater deploys the failed digest again | Resolve and smoke the corrected `prod` digest first. |
| A source revert triggers four legacy image workflows | Use path-scoped Paper workflow and inspect all resulting package events before declaring recovery complete. |
