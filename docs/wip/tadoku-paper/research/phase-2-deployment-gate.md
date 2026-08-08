# Phase 2 deployment gate

Date: 2026-08-08  
Status: passed

## Outcome

The Button, Input, Modal, and ActionMenu vertical slice is live at
`paper.tadoku.app`. The slice proves one shared raw-class/React action recipe,
React Hook Form field anatomy, Base UI focus and composite-keyboard behavior,
deterministic fixtures, and the complete 16-section Stable page contract.

Real-browser review found that Base UI's default portals escaped the isolated
preview iframe. Modal and ActionMenu now derive their portal container from
the trigger's owner document. Regression tests cover both wrappers, and the
review was repeated before merge.

## Release record

| Field | Value |
| --- | --- |
| Source PR | `tadoku/tadoku#783` |
| Tadoku source | `421cac67e7aaf50b4e020c4dd6e13b9983d1f3d5` |
| Incoming image | `ghcr.io/tadoku/tadoku/frontend-paper-styleguide@sha256:a1578d8ca316ab5b4bbac5c172fed94e1e487f4e410da859d624c2c705bc0320` |
| Outgoing image | `ghcr.io/tadoku/tadoku/frontend-paper-styleguide@sha256:6a67d1f77c1b124f79fa1b6e9fe283443517da41de5d40322c444cf0e5e0d0a3` |
| GitOps revision | `4bb9d3fdff59394bcb3a5622666817f89274451e` |
| Application | `tadoku-paper-styleguide` in `tdk-prod-paper-styleguide` |
| Rollback | Pause the Paper updater and pin the outgoing digest |

## Verification

| Gate | Evidence | Result |
| --- | --- | --- |
| Package | Lint, typecheck, 37 tests, build, TypeScript 4.9 consumer, pack check | Pass |
| Styleguide | Lint, typecheck, 11 tests, Vite production build | Pass |
| Coexistence | Complete frontend workspace build; Paper boundaries | Pass |
| Static delivery | Exact image build and hardened container smoke | Pass |
| Instructional schema | Four Stable pages, all 16 required sections, fixture-driven Preview/Code/API/Accessibility views | Pass |
| Browser | Phone and desktop layouts; real hover/focus, menu pointer/keyboard, modal containment/Escape/return, RHF error relationships | Pass |
| Reconciliation | Argo CD `Synced Healthy`; ready pod, exact incoming image ID, zero restarts | Pass |
| Public routes | Button, Input, Modal, and ActionMenu canonical routes returned `200`; `/.env` returned `404` | Pass |
| Legacy independence | `ui.tadoku.app` returned `200` unchanged | Pass |

The temporary `t3-expose` review session was stopped after the production
acceptance sweep.

## Gate decision

Phase 2 passes. Phase 3 may complete the package surface and catalogue. No
production application migration is authorized by this gate.
