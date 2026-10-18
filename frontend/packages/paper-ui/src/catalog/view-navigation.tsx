import Example40 from "./examples/navigation.pagination.entries";
import Example40Source from "./examples/navigation.pagination.entries.tsx?raw";
import Example38 from "./examples/navigation.tabbar.contests";
import Example38Source from "./examples/navigation.tabbar.contests.tsx?raw";
import Example39 from "./examples/navigation.vertical-tabbar.profile";
import Example39Source from "./examples/navigation.vertical-tabbar.profile.tsx?raw";
import { defineCatalogFixture } from "./schema";
import type { NavigationDocSpec } from "./phase-three-navigation";

export const viewNavigationFixtures = [
  defineCatalogFixture({ ...{
  "id": "navigation.tabbar.contests",
  "name": "Contest views",
  "description": "Choose a linked view and observe its current rail. On Phone, scroll the list without hiding destinations.",
  "tags": [
    "navigation",
    "tabbar",
    "contests"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example38Source, render: () => <Example38 /> }),
  defineCatalogFixture({ ...{
  "id": "navigation.vertical-tabbar.profile",
  "name": "Profile views",
  "description": "Choose peer destinations within a bounded navigation column. Parent layout supplies its width.",
  "tags": [
    "navigation",
    "vertical",
    "tabbar",
    "profile"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example39Source, render: () => <Example39 /> }),
  defineCatalogFixture({ ...{
  "id": "navigation.pagination.entries",
  "name": "Reading-entry pages",
  "description": "Move the controlled result page; compare the middle-page window with the single-page boundary.",
  "tags": [
    "navigation",
    "pagination",
    "entries"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example40Source, render: () => <Example40 /> })
] as const;

export const viewNavigationSpecs: readonly NavigationDocSpec[] = [
  {
    id: "component.tabbar",
    route: "/components/navigation/tabbar",
    name: "Tabbar",
    example: "Select Official contests, Community contests or My contests. The example updates currentPath through onSelect and uses real fragment hrefs; replace them with application routes. Phone demonstrates horizontal overflow.",
    props: [{ "name": "links", "type": "readonly TabbarItem[]", "required": true, "description": "Peer destinations with unique id, label and href. Avoid more items than users can compare." }, { "name": "links[].id / label / href / current / disabled / icon / onSelect", "type": "NavigationItem", "description": "Stable id, visible label and destination href; optional current state, disabled state, decorative icon and selection callback. onSelect does not update currentPath for you." }, { "name": "currentPath", "type": "string", "description": "Exact href match for current state. An item\u2019s explicit current boolean takes precedence." }, { "name": "renderLink", "type": "NavigationLinkRenderer", "description": "Optional router adapter returning an anchor. Forward every supplied prop, including handlers and accessibility attributes.", "defaultValue": "native <a>" }, { "name": "label", "type": "string", "description": "Accessible navigation landmark name; distinguish multiple navigation regions." }],
    summary: "Displays peer destination views as horizontal linked navigation that remains scrollable at narrow widths.",
    fixtureId: "navigation.tabbar.contests",
    react: ["Tabbar"],
    types: ["TabbarProps", "TabbarItem", "NavigationLinkRenderer"],
    css: ["paper-tabbar", "paper-tabbar__list--horizontal", "paper-tabbar__link"],
    when: "Use for a small set of peer routes representing different views of the same context.",
    avoid: "Do not apply tab roles to links that navigate and do not hide primary destinations in a menu solely for narrow widths.",
    choose: "Use VerticalTabbar when vertical space and labels fit better; use native in-page tabs only for panels that do not navigate.",
    anatomy: "A named navigation landmark, native list, linked destinations, and a straight current-route lower rail.",
    variants: "Horizontal linked navigation supports current, default, disabled, icon-bearing, and overflow states.",
    states: "The current lower rail stays distinct from keyboard focus. Below the available width the list scrolls horizontally; Tab brings offscreen links into view. Density changes target height, not the number of destinations.",
    behavior: "Links remain in normal Tab order and activate with native link behavior; horizontal overflow remains touch and keyboard scrollable.",
    content: "Use short parallel view labels and keep the number of peer destinations small.",
    accessibility: "Use navigation and aria-current rather than tablist/tab roles because activation changes route.",
    migration: "Replace Next Link and responsive ActionMenu coupling with renderLink and a visible overflow-safe list."
  },
  {
    id: "component.vertical-tabbar",
    route: "/components/navigation/vertical-tabbar",
    name: "VerticalTabbar",
    example: "Select a profile view and watch the left rail follow currentPath. The 22rem wrapper is a layout choice, not part of the component API. Keep the content region beside it on wide layouts and below it on narrow layouts.",
    props: [{ "name": "links", "type": "readonly TabbarItem[]", "required": true, "description": "Peer destinations with unique id, label and href. Width and placement beside content belong to the surrounding layout." }, { "name": "links[].id / label / href / current / disabled / icon / onSelect", "type": "NavigationItem", "description": "Stable id, visible label and destination href; optional current state, disabled state, decorative icon and selection callback. onSelect does not update currentPath for you." }, { "name": "currentPath", "type": "string", "description": "Exact href match for current state. An item\u2019s explicit current boolean takes precedence." }, { "name": "renderLink", "type": "NavigationLinkRenderer", "description": "Optional router adapter returning an anchor. Forward every supplied prop, including handlers and accessibility attributes.", "defaultValue": "native <a>" }, { "name": "label", "type": "string", "description": "Accessible navigation landmark name; distinguish multiple navigation regions." }],
    summary: "Presents peer destination views vertically while preserving linked-navigation semantics.",
    fixtureId: "navigation.vertical-tabbar.profile",
    react: ["VerticalTabbar"],
    types: ["TabbarProps", "TabbarItem", "NavigationLinkRenderer"],
    css: ["paper-tabbar--vertical", "paper-tabbar__list--vertical", "paper-tabbar__link"],
    when: "Use when peer route labels benefit from vertical measure beside their content.",
    avoid: "Do not use for broad multi-section application navigation or non-routing panel controls.",
    choose: "Use Tabbar for compact horizontal peer routes and Sidebar for grouped application areas.",
    anatomy: "A named navigation landmark, vertical list, linked destinations, and a straight current-route side rail.",
    variants: "Vertical linked navigation shares the same current, disabled, icon, and router adapter contract as Tabbar.",
    states: "Current state uses a left rail. The component remains vertical at every breakpoint; the application moves the navigation above content when the side column no longer fits. Use short labels and test localized content.",
    behavior: "Native link focus and activation remain unchanged; no arrow-key tab composite is introduced.",
    content: "Use concise parallel labels describing peer views of one object or task.",
    accessibility: "Expose navigation plus aria-current and do not use tab roles for route changes.",
    migration: "Replace the undocumented legacy VerticalTabbar and Next Link coupling while retaining destination hrefs."
  },
  {
    id: "component.pagination",
    route: "/components/navigation/pagination",
    name: "Pagination",
    example: "Previous, Next and numbered buttons update the displayed page and entry range. Compare the one-page example, whose edge controls are unavailable. Real data fetching, loading announcements and moving focus to results belong to the application.",
    props: [{ "name": "totalPages", "type": "number", "required": true, "description": "Positive integer. Omit the component when there are no result pages." }, { "name": "currentPage", "type": "number", "required": true, "description": "One-based controlled page between 1 and totalPages. Invalid boundaries throw RangeError." }, { "name": "getHref", "type": "(page: number) => string", "description": "Creates native destination anchors. At least getHref or onPageChange is required." }, { "name": "onPageChange", "type": "(page: number) => void", "description": "Update your page state and result data. Alone: buttons. With getHref: intercept unmodified primary activation; modified clicks keep browser behavior." }, { "name": "siblingCount", "type": "number", "defaultValue": "2", "description": "Non-negative integer number of neighboring pages on each side; first and last are retained." }, { "name": "renderLink", "type": "NavigationLinkRenderer", "description": "Optional router adapter returning an anchor. Forward every supplied prop, including handlers and accessibility attributes.", "defaultValue": "native <a>" }, { "name": "label", "type": "string", "defaultValue": "Pagination", "description": "Accessible landmark name. Distinguish multiple result sets." }],
    summary: "Moves through a known bounded page set with href, callback, or combined router integration.",
    fixtureId: "navigation.pagination.entries",
    react: ["Pagination"],
    types: ["PaginationProps", "NavigationLinkRenderer"],
    css: ["paper-pagination", "paper-pagination__control", "paper-pagination__page"],
    when: "Use when a result collection is split into stable numbered pages and direct page URLs or controlled page state exist.",
    avoid: "Do not use for unknown unbounded feeds, carousel slides, or step-by-step forms.",
    choose: "Use getHref for crawlable/direct routes, onPageChange for controlled local state, or both when an app router intercepts anchors.",
    anatomy: "A named nav, native list, previous/next controls, current-page text, bounded numbered destinations, and non-interactive gaps.",
    variants: "Href, callback, and combined modes share siblingCount and compact responsive behavior.",
    states: "Below 48rem the numbered links and gaps are hidden; Previous and Next remain. Pair them with a visible Page X of Y summary, as this example does. At the first and last page the unavailable edge is non-interactive.",
    behavior: "Native anchors preserve modified-click browser behavior. With onPageChange, ordinary activation updates controlled state; callback-only mode uses buttons. Current-page display does not update until the application passes the new currentPage.",
    content: "Keep Previous and Next labels visible and expose each numeric destination as Page plus its number.",
    accessibility: "Name the landmark, identify the current page with aria-current, label page destinations, hide decorative gaps, and remove unavailable edges from interaction.",
    migration: "Replace Next router/query access and the page-jump modal with explicit getHref or onPageChange supplied by the application."
  }
];
