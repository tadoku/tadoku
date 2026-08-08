import type { ReactNode } from "react";
import {
  Breadcrumb,
  Navbar,
  Pagination,
  Sidebar,
  Tabbar,
  VerticalTabbar,
} from "../components/navigation";
import {
  defineCatalogDocument,
  defineCatalogFixture,
  type CatalogDocument,
  type ComponentDocumentationSections,
  type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-08-08";
const VIEWPORTS = [
  { id: "phone", label: "Phone", width: 360, height: 720 },
  { id: "tablet", label: "Tablet", width: 768, height: 800 },
  { id: "desktop", label: "Desktop", width: 1280, height: 800 },
] as const;

const navbarNavigation = [
  { type: "link", id: "home", label: "Home", href: "#home", current: true },
  { type: "link", id: "contests", label: "Contests", href: "#contests" },
  { type: "link", id: "manual", label: "Manual", href: "#manual" },
  {
    type: "dropdown",
    id: "account",
    label: "Account",
    links: [
      { id: "profile", label: "Profile", href: "#profile" },
      { id: "settings", label: "Settings", href: "#settings" },
      { id: "sign-out", label: "Sign out", href: "#sign-out" },
    ],
  },
] as const;

const profileLinks = [
  { id: "entries", label: "Reading entries", href: "#entries" },
  { id: "statistics", label: "Statistics", href: "#statistics", current: true },
  { id: "contests", label: "Contests", href: "#profile-contests" },
] as const;

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
}

function navigationSections(spec: NavigationDocSpec): ComponentDocumentationSections {
  const required: RequiredComponentSections = {
    overview: { heading: "Overview", content: [spec.summary] },
    whenToUse: { heading: "When to use", content: [spec.when] },
    whenNotToUse: { heading: "When not to use", content: [spec.avoid] },
    choosingBetween: { heading: "Choose between", content: [spec.choose] },
    anatomy: { heading: "Anatomy", content: [spec.anatomy] },
    recommendedExample: { heading: "Recommended example", content: [`The ${spec.fixtureId} fixture uses stable Tadoku routes and content without owning a router.`] },
    variants: { heading: "Variants", content: [spec.variants] },
    statesAndAdaptation: { heading: "States and adaptation", content: [spec.states] },
    behavior: { heading: "Behavior", content: [spec.behavior] },
    contentGuidance: { heading: "Content guidance", content: [spec.content] },
    accessibility: { heading: "Accessibility", content: [spec.accessibility] },
    implementation: { heading: "Implementation", content: [`Import ${spec.react.join(", ")} from paper-ui, load paper-ui/styles.css once, and pass application routing through renderLink.`] },
    apiReference: { heading: "API reference", content: [`React: ${spec.react.join(", ")}. Types: ${spec.types.join(", ")}. CSS: ${spec.css.join(", ")}.`] },
    relatedPatterns: { heading: "Related patterns", content: ["Combine navigation components with the application shell and real destination data; do not duplicate routing or authorization policy inside Paper."] },
    migration: { heading: "Migration", content: [spec.migration] },
    lifecycle: { heading: "Lifecycle", content: ["Stable in Paper 0.1.0 with deterministic fixtures and rendered semantic, keyboard, and router-boundary tests."] },
  };
  return { required };
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
      cssClasses: spec.css,
      publicTypes: spec.types,
      defaults: ["Links render as native anchors; renderLink may replace them without changing semantics."],
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

const fixture = (
  id: string,
  name: string,
  description: string,
  code: string,
  render: () => ReactNode,
) => defineCatalogFixture({
  id,
  name,
  description,
  tags: id.split(/[.-]/u),
  themes: ["light", "dark"],
  densities: ["comfortable", "compact"],
  viewports: VIEWPORTS,
  deterministic: true,
  code,
  render,
});

export const phaseThreeNavigationFixtures = [
  fixture(
    "navigation.navbar.primary",
    "Primary Tadoku navigation",
    "Current route, account dropdown, loading state, and narrow disclosure behavior.",
    `import { Navbar } from "paper-ui";\n\n<Navbar brand="Tadoku" brandHref="/" currentPath="/" navigation={navigation} renderLink={renderAppLink} />`,
    () => <Navbar brand="Tadoku" brandHref="#home" navigation={navbarNavigation} />,
  ),
  fixture(
    "navigation.sidebar.admin",
    "Admin sidebar",
    "Grouped administration routes with current and disabled states.",
    `import { Sidebar } from "paper-ui";\n\n<Sidebar label="Admin sections" currentPath="/admin/logs" sections={sections} renderLink={renderAppLink} />`,
    () => <Sidebar label="Admin sections" sections={[{ id: "content", title: "Content", links: [{ id: "logs", label: "Reading logs", href: "#logs", current: true }, { id: "users", label: "Users", href: "#users" }] }, { id: "system", title: "System", links: [{ id: "imports", label: "Imports", href: "#imports" }, { id: "jobs", label: "Background jobs", href: "#jobs", disabled: true }] }]} />,
  ),
  fixture(
    "navigation.breadcrumb.contest",
    "Contest breadcrumb",
    "A location trail from Tadoku home to one contest entry.",
    `import { Breadcrumb } from "paper-ui";\n\n<Breadcrumb items={[{ id: "home", label: "Home", href: "/" }, { id: "contests", label: "Contests", href: "/contests" }, { id: "current", label: "August Japanese" }]} renderLink={renderAppLink} />`,
    () => <Breadcrumb items={[{ id: "home", label: "Home", href: "#home" }, { id: "contests", label: "Contests", href: "#contests" }, { id: "round", label: "August Japanese" }]} />,
  ),
  fixture(
    "navigation.tabbar.contests",
    "Contest views",
    "Horizontally scrollable linked views with one current destination.",
    `import { Tabbar } from "paper-ui";\n\n<Tabbar label="Contest views" links={links} renderLink={renderAppLink} />`,
    () => <Tabbar label="Contest views" links={[{ id: "official", label: "Official contests", href: "#official" }, { id: "community", label: "Community contests", href: "#community" }, { id: "mine", label: "My contests", href: "#mine", current: true }]} />,
  ),
  fixture(
    "navigation.vertical-tabbar.profile",
    "Profile views",
    "Vertical linked navigation for profile content regions.",
    `import { VerticalTabbar } from "paper-ui";\n\n<VerticalTabbar label="Profile views" links={links} renderLink={renderAppLink} />`,
    () => <VerticalTabbar label="Profile views" links={profileLinks} />,
  ),
  fixture(
    "navigation.pagination.entries",
    "Reading-entry pages",
    "A bounded page window with router hrefs, edge controls, and compact adaptation.",
    `import { Pagination } from "paper-ui";\n\n<Pagination totalPages={24} currentPage={8} getHref={(page) => \`/entries?page=\${page}\`} renderLink={renderAppLink} />`,
    () => <Pagination totalPages={24} currentPage={8} getHref={(page) => `#page-${page}`} />,
  ),
] as const;

const specs: readonly NavigationDocSpec[] = [
  {
    id: "component.navbar",
    route: "/components/navigation/navbar",
    name: "Navbar",
    summary: "Provides responsive global navigation, account destinations, current-route state, and progress feedback without owning application routing.",
    fixtureId: "navigation.navbar.primary",
    react: ["Navbar"],
    types: ["NavbarProps", "NavbarItem", "NavigationDirectLink", "NavigationDropdown", "NavigationLinkRenderer"],
    css: ["paper-navbar", "paper-navbar__link", "paper-navbar__dropdown"],
    when: "Use once as the application's primary global navigation landmark.",
    avoid: "Do not use for local page sections, contextual actions, or authorization decisions.",
    choose: "Use Sidebar for grouped local destinations and ActionMenu for contextual commands.",
    anatomy: "A brand link, direct destination links, optional Base UI dropdowns, a narrow disclosure button, and an optional named loading rail.",
    variants: "Direct links and account dropdowns share the same current-path and renderLink contract.",
    states: "Review signed-out, signed-in, admin, banned, loading, open dropdown, open narrow menu, disabled destination, and long-label states.",
    behavior: "Tab reaches links and triggers; Enter or Space opens disclosures; Base UI supplies menu arrows, Home/End, Escape, outside dismissal, and focus return.",
    content: "Use short destination nouns and a specific account-group label; brand content must retain an accessible name.",
    accessibility: "Keep one named navigation landmark, aria-current on the active destination, expanded/control relationships on the narrow trigger, and owner-document portals.",
    migration: "Move Next route comparison and Link rendering into the application adapter; replace Headless UI disclosure/menu imports with Paper Navbar data.",
    packages: ["paper-ui", "@base-ui/react"],
  },
  {
    id: "component.sidebar",
    route: "/components/navigation/sidebar",
    name: "Sidebar",
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
    states: "Review current, hover, focus, disabled, compact density, narrow width, and long localized labels.",
    behavior: "Native links participate in sequential focus; disabled destinations are removed from Tab order and block activation.",
    content: "Section headings name coherent destination groups; link labels use parallel destination nouns.",
    accessibility: "Use unique section IDs, aria-current page, aria-disabled with blocked activation, and icons hidden from the accessible name.",
    migration: "Replace Next Link with renderLink and pass currentPath or explicit current state from the application.",
  },
  {
    id: "component.breadcrumb",
    route: "/components/navigation/breadcrumb",
    name: "Breadcrumb",
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
    states: "Review one-item, deep, narrow, wrapping, long-label, and missing-parent-href trails.",
    behavior: "Parent anchors use native link keyboard behavior; separators and the current item are not interactive.",
    content: "Use the same concise labels as destination headings and omit redundant words such as page.",
    accessibility: "Keep ordered-list structure, hide separators, render one aria-current page, and never disable a current-page anchor as a substitute for text.",
    migration: "Replace legacy Next links and the mobile back-link branch with one semantic responsive trail and renderLink adapter.",
  },
  {
    id: "component.tabbar",
    route: "/components/navigation/tabbar",
    name: "Tabbar",
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
    states: "Review current, focus, disabled, narrow overflow, compact density, and long-label states.",
    behavior: "Links remain in normal Tab order and activate with native link behavior; horizontal overflow remains touch and keyboard scrollable.",
    content: "Use short parallel view labels and keep the number of peer destinations small.",
    accessibility: "Use navigation and aria-current rather than tablist/tab roles because activation changes route.",
    migration: "Replace Next Link and responsive ActionMenu coupling with renderLink and a visible overflow-safe list.",
  },
  {
    id: "component.vertical-tabbar",
    route: "/components/navigation/vertical-tabbar",
    name: "VerticalTabbar",
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
    states: "Review current, focus, disabled, compact, narrow container, and long-label states.",
    behavior: "Native link focus and activation remain unchanged; no arrow-key tab composite is introduced.",
    content: "Use concise parallel labels describing peer views of one object or task.",
    accessibility: "Expose navigation plus aria-current and do not use tab roles for route changes.",
    migration: "Replace the undocumented legacy VerticalTabbar and Next Link coupling while retaining destination hrefs.",
  },
  {
    id: "component.pagination",
    route: "/components/navigation/pagination",
    name: "Pagination",
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
    states: "Review first, middle, last, one-page, gap, compact viewport, and invalid-boundary inputs.",
    behavior: "Native anchors preserve open-in-new-tab behavior unless onPageChange intentionally intercepts; callback-only mode uses buttons; unavailable edges are non-interactive.",
    content: "Keep Previous and Next labels visible and expose each numeric destination as Page plus its number.",
    accessibility: "Name the landmark, identify the current page with aria-current, label page destinations, hide decorative gaps, and remove unavailable edges from interaction.",
    migration: "Replace Next router/query access and the page-jump modal with explicit getHref or onPageChange supplied by the application.",
  },
];

export const phaseThreeNavigationDocuments = specs.map(navigationDocument);
