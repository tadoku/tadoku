import Example37 from "./examples/navigation.breadcrumb.contest";
import Example37Source from "./examples/navigation.breadcrumb.contest.tsx?raw";
import Example35 from "./examples/navigation.navbar.primary";
import Example35Source from "./examples/navigation.navbar.primary.tsx?raw";
import Example36 from "./examples/navigation.sidebar.admin";
import Example36Source from "./examples/navigation.sidebar.admin.tsx?raw";
import { defineCatalogFixture } from "./schema";
import type { NavigationDocSpec } from "./phase-three-navigation";

export const siteNavigationFixtures = [
  defineCatalogFixture({ ...{
  "id": "navigation.navbar.primary",
  "name": "Primary Tadoku navigation",
  "description": "Choose a destination, open Guides, or switch to Phone and open Menu. Your profile appears at the bottom of the mobile menu. A second bar demonstrates loading.",
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
] as const;

export const siteNavigationSpecs: readonly NavigationDocSpec[] = [
  {
    id: "component.navbar",
    route: "/components/navigation/navbar",
    name: "Navbar",
    example: "Choose a destination to see the active route change. Open Guides on Desktop; choose Phone and open Menu to see the same destinations plus Your profile at the bottom. The mobileFooter callback supplies closeMenu so account links and commands can close the disclosure after selection. A separate loading bar illustrates progress without starting a timer.",
    props: [{ "name": "brand / brandHref", "type": "ReactNode / string", "required": true, "description": "Accessible brand content and its home destination." }, { "name": "navigation", "type": "readonly NavbarItem[]", "required": true, "description": 'Direct {type:"link", id, label, href} destinations or {type:"dropdown", id, label, links} destination groups.' }, { "name": "actions", "type": "ReactNode", "description": "End-aligned controls visible at every breakpoint; keep them short enough to fit beside the brand and mobile menu trigger." }, { "name": "mobileFooter", "type": "(closeMenu: () => void) => ReactNode", "description": "Optional account links or commands at the bottom of the mobile disclosure. Call closeMenu after choosing a destination or completing an action; this content is hidden on desktop." }, { "name": "mobileNavigation", "type": "boolean", "defaultValue": "true", "description": "Show disclosure below 48rem. Disable only when the shell supplies an alternate narrow navigation." }, { "name": "menuLabel", "type": "string", "defaultValue": "Menu", "description": "Used in the narrow trigger accessible name: Open menu / Close menu." }, { "name": "isLoading", "type": "boolean", "defaultValue": "false", "description": "Show the named indeterminate route-loading rail; the application controls its duration." }, { "name": "currentPath", "type": "string", "description": "Exact href match for current state. An item\u2019s explicit current boolean takes precedence." }, { "name": "renderLink", "type": "NavigationLinkRenderer", "description": "Optional router adapter returning an anchor. Forward every supplied prop, including handlers and accessibility attributes.", "defaultValue": "native <a>" }, { "name": "label", "type": "string", "description": "Accessible navigation landmark name; distinguish multiple navigation regions." }],
    summary: "Provides responsive global navigation, account destinations, current-route state, and progress feedback without owning application routing.",
    fixtureId: "navigation.navbar.primary",
    react: ["Navbar"],
    types: ["NavbarProps", "NavbarItem", "NavigationDirectLink", "NavigationDropdown", "NavigationLinkRenderer"],
    css: ["paper-navbar", "paper-navbar__link", "paper-navbar__dropdown", "paper-navbar__mobile-link"],
    when: "Use once as the application's primary global navigation landmark.",
    avoid: "Do not use for local page sections, contextual actions, or authorization decisions.",
    choose: "Use Sidebar for grouped local destinations and ActionMenu for contextual commands.",
    anatomy: "A brand link, direct destination links, optional Base UI dropdowns, an optional end-actions slot, a narrow disclosure with an optional mobile footer, and an optional named loading rail.",
    variants: "Direct links and account dropdowns share the same current-path and renderLink contract. Set mobileNavigation to false when an application shell supplies a Drawer or another narrow navigation pattern.",
    states: "Below 48rem, direct links and dropdown destinations move into the Menu disclosure, followed by mobileFooter when supplied. Use this footer for account links or commands that belong inside the mobile menu. The application decides which links exist for each account role. Loading adds a status rail without disabling navigation.",
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
];
