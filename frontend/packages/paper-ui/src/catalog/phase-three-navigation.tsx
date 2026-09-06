import Example37 from "./examples/navigation.breadcrumb.contest";
import Example37Source from "./examples/navigation.breadcrumb.contest.tsx?raw";
import Example35 from "./examples/navigation.navbar.primary";
import Example35Source from "./examples/navigation.navbar.primary.tsx?raw";
import Example40 from "./examples/navigation.pagination.entries";
import Example40Source from "./examples/navigation.pagination.entries.tsx?raw";
import Example36 from "./examples/navigation.sidebar.admin";
import Example36Source from "./examples/navigation.sidebar.admin.tsx?raw";
import Example38 from "./examples/navigation.tabbar.contests";
import Example38Source from "./examples/navigation.tabbar.contests.tsx?raw";
import Example39 from "./examples/navigation.vertical-tabbar.profile";
import Example39Source from "./examples/navigation.vertical-tabbar.profile.tsx?raw";
import {
COMPONENT_PAGE_SECTION_KEYS,
defineCatalogDocument,
defineCatalogFixture,
type CatalogDocument,
type ComponentDocumentationSections,
type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-09-05";

interface NavigationDocSpec {
  readonly id: string;
  readonly route: string;
  readonly name: string;
  readonly summary: string;
  readonly fixtureId: string;
  readonly react: readonly string[];
  readonly types: readonly string[];
  readonly css: readonly string[];
  readonly when: string;
  readonly avoid: string;
  readonly choose: string;
  readonly anatomy: string;
  readonly variants: string;
  readonly states: string;
  readonly behavior: string;
  readonly content: string;
  readonly accessibility: string;
  readonly migration: string;
  readonly packages?: readonly string[];
  readonly example: string;
  readonly props: NonNullable<CatalogDocument["api"]["props"]>;
}

function navigationSections(spec: NavigationDocSpec): ComponentDocumentationSections {
  const required: RequiredComponentSections = {
    overview: { heading: "Overview", content: [spec.summary] },
    whenToUse: { heading: "When to use", content: [spec.when] },
    whenNotToUse: { heading: "When not to use", content: [spec.avoid] },
    choosingBetween: { heading: "Choose between", content: [spec.choose] },
    anatomy: { heading: "Anatomy", content: [spec.anatomy] },
    recommendedExample: { heading: "Recommended example", content: [spec.example] },
    variants: { heading: "Variants", content: [spec.variants] },
    statesAndAdaptation: { heading: "States and adaptation", content: [spec.states] },
    behavior: { heading: "Behavior", content: [spec.behavior] },
    contentGuidance: { heading: "Content guidance", content: [spec.content] },
    accessibility: { heading: "Accessibility", content: [spec.accessibility] },
    implementation: { heading: "Implementation", content: [`Import ${spec.react.join(", ")} from paper-ui and load paper-ui/styles.css once at the application entry. Examples use native fragment links so they run without a router. For an application router, pass renderLink={(props) => <AppLink {...props} />}; forward href, className, children, tabIndex, event handlers and aria attributes to the final anchor. Derive currentPath from your router; current on an item explicitly overrides exact href matching.`] },
    apiReference: { heading: "API reference", content: [`See API / Props for required data, defaults, and routing contracts. Public types: ${spec.types.join(", ")}.`] },
    relatedPatterns: { heading: "Related patterns", content: ["Combine navigation components with the application shell and real destination data; do not duplicate routing or authorization policy inside Paper."] },
    migration: { heading: "Migration", content: [spec.migration] },
    lifecycle: { heading: "Lifecycle", content: ["Stable in Paper 0.1.0 with deterministic fixtures and rendered semantic, keyboard, and router-boundary tests."] },
  };
  return { required, pageSections: COMPONENT_PAGE_SECTION_KEYS };
}

function navigationDocument(spec: NavigationDocSpec): CatalogDocument {
  return defineCatalogDocument({
    id: spec.id,
    route: spec.route,
    name: spec.name,
    kind: "component",
    category: "navigation",
    aliases: [],
    summary: spec.summary,
    keywords: [spec.name.toLocaleLowerCase(), "navigation", "router neutral", "paper"],
    lifecycle: "Stable",
    reviewDate: REVIEW_DATE,
    sourcePath: "src/components/navigation/navigation.tsx",
    packageVersion: "0.1.0",
    guidance: {
      whenToUse: [spec.when],
      whenNotToUse: [spec.avoid],
      content: [spec.content],
      commonMistakes: [spec.choose],
    },
    accessibility: {
      requirements: [spec.accessibility],
      keyboard: [spec.behavior],
      knownConstraints: [spec.states],
    },
    api: {
      react: spec.react,
      props: spec.props,
      cssClasses: spec.css,
      publicTypes: spec.types,
      defaults: [
        "Links render as native anchors; renderLink may replace them without changing semantics.",
        ...(spec.name === "Navbar"
          ? ["actions renders end-aligned utilities; mobileNavigation defaults to true and may be disabled when the application provides another narrow navigation pattern."]
          : []),
      ],
      invalidCombinations: [spec.avoid],
    },
    fixtureIds: [spec.fixtureId],
    dependencies: { documents: ["foundation.iconography"], packages: spec.packages ?? ["paper-ui"] },
    migration: { legacy: [`ui ${spec.name}`], notes: [spec.migration] },
    changelog: [{ date: REVIEW_DATE, note: `Published router-neutral ${spec.name} as Stable.` }],
    behaviorTestIds: [`${spec.id}.semantics`, `${spec.id}.keyboard`, `${spec.id}.router-boundary`],
    sections: navigationSections(spec),
  });
}

export const phaseThreeNavigationFixtures = [
  defineCatalogFixture({ ...{
  "id": "navigation.navbar.primary",
  "name": "Primary Tadoku navigation",
  "description": "Choose a destination, open Account, or switch to Phone and open Menu. A second bar demonstrates loading.",
  "tags": [
    "navigation",
    "navbar",
    "primary"
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
}, code: Example35Source, render: () => <Example35 /> }),
  defineCatalogFixture({ ...{
  "id": "navigation.sidebar.admin",
  "name": "Admin sidebar",
  "description": "Switch current destination within grouped links; Background jobs demonstrates an explained unavailable destination.",
  "tags": [
    "navigation",
    "sidebar",
    "admin"
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
}, code: Example36Source, render: () => <Example36 /> }),
  defineCatalogFixture({ ...{
  "id": "navigation.breadcrumb.contest",
  "name": "Contest breadcrumb",
  "description": "A four-level trail collapses to parent and current location below 48rem. The current page is plain text.",
  "tags": [
    "navigation",
    "breadcrumb",
    "contest"
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
}, code: Example37Source, render: () => <Example37 /> }),
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

const specs: readonly NavigationDocSpec[] = [
  {
    id: "component.navbar",
    route: "/components/navigation/navbar",
    name: "Navbar",
    example: "Choose a destination to see the active route change. Open Account on Desktop; choose Phone and open Menu to see the same destinations as a disclosure. A separate loading bar illustrates progress without starting a timer.",
    props: [{ "name": "brand / brandHref", "type": "ReactNode / string", "required": true, "description": "Accessible brand content and its home destination." }, { "name": "navigation", "type": "readonly NavbarItem[]", "required": true, "description": 'Direct {type:"link", id, label, href} destinations or {type:"dropdown", id, label, links} destination groups.' }, { "name": "actions", "type": "ReactNode", "description": "End-aligned controls; ensure they still fit next to the brand on mobile." }, { "name": "mobileNavigation", "type": "boolean", "defaultValue": "true", "description": "Show disclosure below 48rem. Disable only when the shell supplies an alternate narrow navigation." }, { "name": "menuLabel", "type": "string", "defaultValue": "Menu", "description": "Used in the narrow trigger accessible name: Open menu / Close menu." }, { "name": "isLoading", "type": "boolean", "defaultValue": "false", "description": "Show the named indeterminate route-loading rail; the application controls its duration." }, { "name": "currentPath", "type": "string", "description": "Exact href match for current state. An item\u2019s explicit current boolean takes precedence." }, { "name": "renderLink", "type": "NavigationLinkRenderer", "description": "Optional router adapter returning an anchor. Forward every supplied prop, including handlers and accessibility attributes.", "defaultValue": "native <a>" }, { "name": "label", "type": "string", "description": "Accessible navigation landmark name; distinguish multiple navigation regions." }],
    summary: "Provides responsive global navigation, account destinations, current-route state, and progress feedback without owning application routing.",
    fixtureId: "navigation.navbar.primary",
    react: ["Navbar"],
    types: ["NavbarProps", "NavbarItem", "NavigationDirectLink", "NavigationDropdown", "NavigationLinkRenderer"],
    css: ["paper-navbar", "paper-navbar__link", "paper-navbar__dropdown"],
    when: "Use once as the application's primary global navigation landmark.",
    avoid: "Do not use for local page sections, contextual actions, or authorization decisions.",
    choose: "Use Sidebar for grouped local destinations and ActionMenu for contextual commands.",
    anatomy: "A brand link, direct destination links, optional Base UI dropdowns, an optional end-actions slot, a narrow disclosure button, and an optional named loading rail.",
    variants: "Direct links and account dropdowns share the same current-path and renderLink contract. Set mobileNavigation to false when an application shell supplies a Drawer or another narrow navigation pattern.",
    states: "Below 48rem, direct links and dropdown destinations move into the Menu disclosure. The application decides which links exist for each account role. Loading adds a status rail without disabling navigation.",
    behavior: "Tab reaches links, end actions, and triggers; Enter or Space opens disclosures; Base UI supplies menu arrows, Home/End, Escape, outside dismissal, and focus return.",
    content: "Use short destination nouns and a specific account-group label; brand content must retain an accessible name.",
    accessibility: "Keep one named navigation landmark, aria-current on the active destination, expanded/control relationships on the narrow trigger, and owner-document portals.",
    migration: "Move Next route comparison and Link rendering into the application adapter; replace Headless UI disclosure/menu imports with Paper Navbar data.",
    packages: ["paper-ui", "@base-ui/react"]
  },
  {
    id: "component.sidebar",
    route: "/components/navigation/sidebar",
    name: "Sidebar",
    example: "Choose Reading logs, Users or Imports to move the current rail. Background jobs stays unavailable with an explanation. The example bounds the sidebar to 22rem; position it beside page content in the application shell.",
    props: [{ "name": "sections", "type": "readonly SidebarSection[]", "required": true, "description": "Groups with unique id, visible title and NavigationItem[] links." }, { "name": "links[].id / label / href / current / disabled / icon / onSelect", "type": "NavigationItem", "description": "Stable id, visible label and destination href; optional current state, disabled state, decorative icon and selection callback. onSelect does not update currentPath for you." }, { "name": "currentPath", "type": "string", "description": "Exact href match for current state. An item\u2019s explicit current boolean takes precedence." }, { "name": "renderLink", "type": "NavigationLinkRenderer", "description": "Optional router adapter returning an anchor. Forward every supplied prop, including handlers and accessibility attributes.", "defaultValue": "native <a>" }, { "name": "label", "type": "string", "description": "Accessible navigation landmark name; distinguish multiple navigation regions." }],
    summary: "Groups local application destinations under visible section headings with a straight current-route accent rail.",
    fixtureId: "navigation.sidebar.admin",
    react: ["Sidebar"],
    types: ["SidebarProps", "SidebarSection", "NavigationItem", "NavigationLinkRenderer"],
    css: ["paper-sidebar", "paper-sidebar__section", "paper-sidebar__link"],
    when: "Use for persistent local navigation with several meaningfully named groups.",
    avoid: "Do not use as a second global Navbar or for commands that do not navigate.",
    choose: "Use VerticalTabbar for peer views of one object and Sidebar for broader grouped destinations.",
    anatomy: "A named navigation landmark containing labelled sections, native lists, and router-rendered anchors.",
    variants: "Sections may contain current, default, icon-bearing, and unavailable destinations.",
    states: "Current destinations have a side rail and stronger text. Disabled destinations remain visible but cannot activate. Comfortable and compact density inherit the corresponding control height; place the Sidebar above content or in a Drawer when a side column no longer fits.",
    behavior: "Native links participate in sequential focus; disabled destinations are removed from Tab order and block activation.",
    content: "Section headings name coherent destination groups; link labels use parallel destination nouns.",
    accessibility: "Use unique section IDs, aria-current page, aria-disabled with blocked activation, and icons hidden from the accessible name.",
    migration: "Replace Next Link with renderLink and pass currentPath or explicit current state from the application."
  },
  {
    id: "component.breadcrumb",
    route: "/components/navigation/breadcrumb",
    name: "Breadcrumb",
    example: "Follow the four-level contest trail. Below 48rem only the immediate parent and current entry remain. The final item stays plain text; hrefs on ancestors preserve native navigation.",
    props: [{ "name": "items", "type": "readonly BreadcrumbItem[]", "required": true, "description": "Ordered {id, label, href?, icon?} ancestors followed by the current page. Last item always renders as current text, even if href is supplied." }, { "name": "renderLink", "type": "NavigationLinkRenderer", "description": "Optional router adapter returning an anchor. Forward every supplied prop, including handlers and accessibility attributes.", "defaultValue": "native <a>" }, { "name": "label", "type": "string", "defaultValue": "Breadcrumb", "description": "Accessible name for this location trail." }],
    summary: "Communicates hierarchical location with an ordered trail whose current page is not an unnecessary link.",
    fixtureId: "navigation.breadcrumb.contest",
    react: ["Breadcrumb"],
    types: ["BreadcrumbProps", "BreadcrumbItem", "NavigationLinkRenderer"],
    css: ["paper-breadcrumb", "paper-breadcrumb__list", "paper-breadcrumb__current"],
    when: "Use when users can arrive deep in a hierarchy and benefit from parent destinations.",
    avoid: "Do not mirror flat history, workflow steps, or primary navigation.",
    choose: "Breadcrumb shows hierarchy; Pagination moves through result pages and Tabbar switches peer views.",
    anatomy: "A named nav, ordered list, parent links, hidden separator icons, and one aria-current terminal item.",
    variants: "Items may carry decorative icons; narrow layouts retain the parent and current location.",
    states: "Below 48rem, all but the immediate parent and current location are hidden. A single item remains current text. Missing ancestor hrefs render as text; supply hrefs when users can navigate upward.",
    behavior: "Parent anchors use native link keyboard behavior; separators and the current item are not interactive.",
    content: "Use the same concise labels as destination headings and omit redundant words such as page.",
    accessibility: "Keep ordered-list structure, hide separators, render one aria-current page, and never disable a current-page anchor as a substitute for text.",
    migration: "Replace legacy Next links and the mobile back-link branch with one semantic responsive trail and renderLink adapter."
  },
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

export const phaseThreeNavigationDocuments = specs.map(navigationDocument);
