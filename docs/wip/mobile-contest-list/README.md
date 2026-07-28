# Mobile contest list

## Status

Planning only. The design and implementation can be refined before work begins.

## Problem

Contest lists use the desktop table layout at every viewport size. On narrow
screens, contest titles and full dates wrap into tall columns, making each row
hard to scan. The page header also presents two dropdowns—contest type and
actions—while the breadcrumb, announcement, and footer compete with the list
for limited vertical space.

The same `ContestList` component is used by official, user, and personal
contest pages. User contest rows can contain two additional columns for
languages and activities, so a mobile treatment must work for both the
three-column and five-column variants.

## Goals

- Make contests easy to scan and open on phone-sized screens.
- Keep the existing desktop table behavior.
- Give the primary create action a clear label instead of hiding it behind a
  generic menu when it is the only available action.
- Make switching between official, user, and personal contests obvious.
- Apply the same responsive behavior to all three contest-list pages.
- Use the existing `ui` package and its button styles.

## Non-goals

- Redesigning contest detail or leaderboard pages.
- Changing contest-list API responses, filtering, or pagination.
- Redesigning the global navigation, announcement system, or footer.
- Adding new contest lifecycle behavior to the backend.

## Proposed mobile experience

At widths below the `md` breakpoint:

1. Replace the table with a vertical list of cards.
2. Show the contest title as the card heading.
3. Present the start and end dates as one compact date range.
4. Make the whole card a clear navigation target and include a trailing
   chevron.
5. Preserve the private-contest indicator and its accessible description.
6. For user contests, place language and activity restrictions below the date
   range only when those values are available.
7. Use concise, always-visible contest-type controls such as `Official`,
   `User`, and `Mine`, while keeping unambiguous accessible labels.
8. Show a direct `Create contest` or `Create` button when creation is allowed
   and it is the only page action.
9. Hide the redundant breadcrumb on mobile, or reduce it to a compact back
   link if user testing shows that it is still useful.

At `md` and wider, retain the current table and full tab bar. The mobile and
desktop views should render from the same contest data and use the same routes.

An optional follow-up can add `Active`, `Upcoming`, and `Ended` badges. If
added, status calculation should be defined once and tested around start/end
boundaries.

## Implementation plan

1. Add focused rendering tests for `ContestList` covering an official contest,
   a private contest, a user contest with restrictions, and an empty result.
2. Extract a small contest date-range formatter so cards and tests do not
   duplicate Luxon formatting details.
3. Add a mobile card view to
   `frontend/apps/webv2/app/immersion/ContestList.tsx`, visible below `md`.
4. Keep the existing table in the same component and show it from `md`
   upwards.
5. Update the three contest-list pages to use compact mobile contest-type
   navigation and a direct create button where appropriate:
   - `pages/contests/official/[[...page]].tsx`
   - `pages/contests/user-contests/[[...page]].tsx`
   - `pages/contests/my-contests/[[...page]].tsx`
6. Condense or hide the contest breadcrumb below `md` without changing
   breadcrumbs elsewhere in the application.
7. Verify empty, loading, error, single-item, multi-item, private, and paginated
   states at representative mobile and desktop widths.
8. Run the webv2 typecheck, lint, and frontend build.

If the direct single-action behavior or compact tab treatment is useful across
the application, add it to the `ui` package as an explicit component option.
Otherwise, keep the behavior scoped to the contest-list pages rather than
changing every existing `ButtonGroup` or `Tabbar` consumer.

## Accessibility

- Each card must have one clear accessible name and navigation destination.
- Do not create nested interactive elements inside the card link.
- Keep private-contest context available to screen readers.
- Ensure the selected contest type is programmatically identifiable.
- Maintain visible keyboard focus and at least a 44px touch target for
  navigation and create actions.
- Do not rely on color alone for optional lifecycle statuses.

## Acceptance criteria

- At 320px width, no contest content requires horizontal scrolling.
- Contest titles and date ranges remain readable without table-style column
  wrapping.
- Opening a contest requires a single obvious tap.
- Official, user, and personal lists share the same mobile presentation.
- Language and activity restrictions remain discoverable on user contests.
- Private contests retain a visible and accessible indicator.
- Pagination remains usable on mobile.
- The desktop table layout is visually and functionally unchanged.
- Typecheck, lint, and build complete successfully.

## Suggested validation

- Test widths: 320px, 375px, 430px, 768px, and a desktop viewport.
- Test long contest names and long translated date values.
- Test signed-out users, users without create permission, and users with create
  permission.
- Test VoiceOver or another screen reader plus keyboard-only navigation.
- Check the page both with and without an active announcement banner.
