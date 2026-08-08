# Per-application migration risk and smoke matrices

These matrices are pre-cutover checklists. Each smoke run needs desktop and narrow layouts, a recorded immutable image/digest, a known rollback image, and a zero-mixing audit.

## Admin (compact density)

### Risk matrix

| Area | Risk | Severity | Evidence / required fixture |
| --- | --- | ---: | --- |
| session/role shell | redirects, access denied, loading, sidebar/mobile shell regress | high | admin session roles; `DashboardLayout`, `AccessDenied`, `LoadingScreen` |
| content editors | CodeMirror/RHF integration, validation, preview state, long content | high | post/page create/edit/preview; app-owned CodeEditor via `useController()` |
| destructive overlays | focus/return and wrong-item deletion/ban | high | content delete, announcement delete, user ban/unban dialogs |
| tables/actions | compact rows, overflow, menu keyboard behavior, pagination | high | content, announcement, language, user lists at 36px |
| announcements | namespace routing and editor/preview state | medium-high | list/new/edit per namespace |
| language/user editing | autocomplete, modal forms, loading/errors | high | languages CRUD; user search/action menu/ban note |
| implicit CSS | global typography, cards, auto-format, table defaults, raw buttons | high | selector inventory and zero-legacy-token report |
| narrow layout | editors/tables/sidebar may overflow or reduce targets | high | phone-width screenshots and keyboard targets |

### Smoke checklist

- [ ] Signed-out request redirects to Auth and return URL remains correct.
- [ ] Signed-in non-admin receives Access Denied and can return to main app.
- [ ] Admin dashboard loads official contest/yearly leaderboard/user totals and loading/error states remain distinguishable.
- [ ] Sidebar opens/closes at narrow width; every section route works; focus and scroll are restored.
- [ ] Posts and Pages: namespace selection, list pagination, create, validation error, edit, preview/version state, delete/cancel.
- [ ] Announcements: namespace list, create, edit, preview, delete/cancel.
- [ ] Languages: search/autocomplete, add/edit state, validation, save/cancel.
- [ ] Users: search/clear, pagination, action menu keyboard operation, ban/unban confirmation and note validation.
- [ ] All dialogs trap/return focus, close with Escape where allowed, and destructive action remains visually/semantically distinct.
- [ ] Tables/editor layouts have no horizontal-page overflow at narrow width; intentional table scrollers work.
- [ ] App lint, typecheck, production build, image build, deployment health, telemetry, and rollback command pass.

## Auth (comfortable density)

### Risk matrix

| Area | Risk | Severity | Evidence / required fixture |
| --- | --- | ---: | --- |
| dynamic Ory nodes | server-defined node types/attributes lose semantics or values | critical | fixtures for input, hidden, submit, button, checkbox, anchor, image, script, text |
| identity methods | OIDC, passkey/WebAuthn, scripts are accidentally filtered/reordered | critical | configured production-like flows and Ory script nodes |
| errors | global and field errors lose name/description association | critical | 400 validation flow updates and global messages |
| submit hierarchy | adjacent-primary CSS hack currently changes second action by DOM order | high | multiple-submit Ory flow; semantic action mapping |
| double submit/loading | repeated credentials/settings submission | high | `aria-busy`, native disabled, keyboard submit fixture |
| navigation/session | signed-in settings vs signed-out flows and return URLs | high | auth Navbar and session redirects |
| duplicate CSS | globals loaded twice today | high | exactly-one Paper stylesheet guard |

### Smoke checklist

- [ ] Login: password path, invalid field/global errors, OIDC action(s), passkey/WebAuthn path, return URL, loading/double-submit prevention.
- [ ] Registration: standard fields, checkbox/terms node if configured, OIDC/passkey path, validation errors, post-registration route.
- [ ] Recovery and verification: initialize/fetch flow, valid submission, expired/invalid flow error recovery, link back to login.
- [ ] Settings (`/`): signed-in profile update, password change, any configured identity method, field/global feedback.
- [ ] Signed-in access to login/recovery redirects as expected; expired/refresh/AAL2/browser-location errors retain current behavior.
- [ ] Every Ory node kind preserves `name`, `value`, required/disabled state, label, hint, message, and script/image/anchor behavior.
- [ ] Tab/Shift-Tab order, Enter submit, focus on error, and accessible error relationships pass.
- [ ] Navbar desktop/mobile states and logout work; external return links remain correct.
- [ ] App lint, typecheck, production build, image build, focused Ory smoke, telemetry, and rollback command pass.

## webv2 (comfortable density)

### Risk matrix

| Area | Risk | Severity | Evidence / required fixture |
| --- | --- | ---: | --- |
| global shell/roles | signed-out/in/admin/banned navigation diverges; mobile disclosure/announcement fails | critical | role fixtures for Navigation, AnnouncementBanner, BannedScreen, Footer |
| dual logging flows | old/new/v2 forms or edit/detail/contest attachment regress | critical | keep all implementations; no product consolidation |
| contests | create/register/leaderboard/updates/profile/actions and role state | critical | official/user/my contest lists and contest role fixtures |
| browse/data | tables, menus, tabs, pagination, dynamic route links | high | all leaderboard/profile/blog list variants |
| charts | palette, heatmap, responsive canvas/overflow | high | activity/split/heatmap fixtures including empty/long data |
| rich content | sanitized blog/pages/manual rely on `.auto-format` and tables | high | prose, inline code, links, multi-table manual |
| page counter | intentionally oversized button is outside standard control sizing | medium-high | documented app exception; touch/keyboard smoke |
| router adapters | Navbar/Breadcrumb/Tabs/Pagination currently rely on Next-aware legacy code | high | app-owned current-route/link adapter tests |
| responsive breadth | widest app surface and table/profile overflow | high | phone/desktop route matrix |

### Smoke checklist

- [ ] Home signed out and signed in; navigation states for user/admin/banned; mobile disclosure, logout, announcement dismiss/action, Footer links.
- [ ] New log old/current v2 forms: all tracking modes, AmountWithUnit, select/grouped select, autocomplete/tags, validation, submit/loading.
- [ ] Log details, edit, delete confirmation, and submit/detach-to-contest flows.
- [ ] Contest lists (official/user/my), create, registration, leaderboard, updates, contest profile; correct available actions by role/state.
- [ ] Latest/all-time/yearly leaderboards: tabs/vertical tabs, ButtonGroup, pagination, narrow table overflow.
- [ ] Profiles: updates, statistics/year switching, heatmap, activity/split charts, contest list, empty/loading/error states.
- [ ] Blog list/post, dynamic content page, manual tables/prose, sanitized links and inline code.
- [ ] Page counter increment/reset behavior with keyboard, touch target, and deliberate oversized visual exception.
- [ ] Every menu/dialog supports keyboard/focus/dismissal; destructive operations are explicit.
- [ ] Direct dynamic routes and paginated deep links preserve path/query/current state.
- [ ] App lint, typecheck, production build, image build, whole-frontend matrix, deployment telemetry, and rollback command pass.

## Paper styleguide pre-migration delivery smoke

Although not a product application cutover, `paper-styleguide` needs its own risk gate: direct nested requests and refresh must return the SPA; search and mobile docs navigation must be keyboard operable; theme/density/viewport controls must affect isolated fixtures; all registry routes/assets/fonts/history links must work; and no `ui`, Next, or Headless UI import/dependency may exist.

## Shared cutover evidence template

Record for each deployment:

- source commit and PR(s);
- new immutable image tag and digest;
- previous immutable image tag and digest;
- exact deploy and rollback commands;
- zero-mixing check output;
- lint/typecheck/test/build/image results;
- smoke operator/time/environment and narrow/desktop coverage;
- health/telemetry observation window and result.

