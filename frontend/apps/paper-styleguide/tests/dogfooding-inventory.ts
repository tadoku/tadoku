export type DogfoodingDebtKind =
  | 'manual-focus-trap'
  | 'manual-tablist'
  | 'native-button'
  | 'native-input'
  | 'native-select'
  | 'surface-style'

export interface DogfoodingDebt {
  readonly key: string
  readonly context: string
  readonly destination: string
}

/**
 * Checked migration ledger for app-owned UI in the Paper styleguide.
 *
 * The companion contract test derives the keys from source. Migrating or
 * adding an item therefore requires an intentional ledger update. Keep the
 * destination phrased as a public Paper component or an explicit Paper API
 * gap; private paper-ui source paths are never a destination.
 */
export const DOGFOODING_DEBT: readonly DogfoodingDebt[] = [
  {
    key: 'src/app/CatalogueSearch.tsx:native-button#1',
    context: 'Search dialog trigger',
    destination: 'Paper Button (outline)',
  },
  {
    key: 'src/app/CatalogueSearch.tsx:native-button#2',
    context: 'Search dialog close action',
    destination: 'Paper Button icon-only API',
  },
  {
    key: 'src/app/DocsShell.tsx:native-button#1',
    context: 'Responsive catalogue navigation trigger',
    destination: 'Paper Button (outline)',
  },
  {
    key: 'src/app/DocsShell.tsx:native-button#2',
    context: 'Responsive catalogue navigation close action',
    destination: 'Paper Button icon-only API',
  },
  {
    key: 'src/documentation/ComponentWorkbench.tsx:native-button#1',
    context: 'Workbench view tabs',
    destination: 'Paper Tabbar',
  },
  {
    key: 'src/documentation/ComponentWorkbench.tsx:native-button#2',
    context: 'Copy-code action',
    destination: 'Paper Button (outline)',
  },
  {
    key: 'src/documentation/ExampleCanvas.tsx:native-button#1',
    context: 'Preview viewport choices',
    destination: 'Paper Tabbar or segmented-choice API',
  },
  {
    key: 'src/app/CatalogueSearch.tsx:native-input#1',
    context: 'Catalogue search query',
    destination: 'Paper Input with react-hook-form',
  },
  {
    key: 'src/documentation/ComponentWorkbench.tsx:native-select#1',
    context: 'Workbench fixture picker',
    destination: 'Paper Select with react-hook-form',
  },
  {
    key: 'src/documentation/ExampleCanvas.tsx:native-select#1',
    context: 'Preview theme picker',
    destination: 'Paper Select with react-hook-form',
  },
  {
    key: 'src/documentation/ExampleCanvas.tsx:native-select#2',
    context: 'Preview density picker',
    destination: 'Paper Select with react-hook-form',
  },
  {
    key: 'src/documentation/ComponentWorkbench.tsx:manual-tablist#1',
    context: 'Workbench view keyboard and selection model',
    destination: 'Paper Tabbar',
  },
  {
    key: 'src/app/CatalogueSearch.tsx:manual-focus-trap#1',
    context: 'Search dialog focus containment',
    destination: 'Paper Modal controlled-open API',
  },
  {
    key: 'src/app/DocsShell.tsx:manual-focus-trap#1',
    context: 'Responsive catalogue drawer focus containment',
    destination: 'Paper drawer primitive built on the public overlay API',
  },
  {
    key: 'src/styles/shell.css:surface-style:.canvas-controls select',
    context: 'Preview preference control surface',
    destination: 'Paper Select recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.canvas-viewport-button',
    context: 'Preview viewport action surface',
    destination: 'Paper Tabbar or segmented-choice recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.catalogue-nav__link--active',
    context: 'Current catalogue navigation item surface',
    destination: 'Paper Sidebar current-item recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.catalogue-nav__link:hover',
    context: 'Hovered catalogue navigation item surface',
    destination: 'Paper Sidebar hover recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.code-view pre',
    context: 'Source-code display surface',
    destination: 'Paper code-block or Surface recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.component-workbench',
    context: 'Examples workbench outer surface',
    destination: 'Paper Surface',
  },
  {
    key: 'src/styles/shell.css:surface-style:.component-workbench__fixture-select select',
    context: 'Fixture picker surface',
    destination: 'Paper Select recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.component-workbench__tabs button:hover',
    context: 'Workbench tab hover surface',
    destination: 'Paper Tabbar hover recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.design-history__links a',
    context: 'Design-history card surface',
    destination: 'Paper Surface with linked-card API',
  },
  {
    key: 'src/styles/shell.css:surface-style:.docs-header',
    context: 'Styleguide header surface',
    destination: 'Paper Navbar',
  },
  {
    key: 'src/styles/shell.css:surface-style:.docs-shell',
    context: 'Styleguide application canvas',
    destination: 'Paper application-shell layout API',
  },
  {
    key: 'src/styles/shell.css:surface-style:.docs-sidebar',
    context: 'Catalogue sidebar surface',
    destination: 'Paper Sidebar',
  },
  {
    key: 'src/styles/shell.css:surface-style:.example-canvas',
    context: 'Isolated preview content surface',
    destination: 'Paper Surface without its own enclosing border',
  },
  {
    key: 'src/styles/shell.css:surface-style:.example-canvas iframe',
    context: 'Isolated preview document surface',
    destination: 'Paper isolated-preview frame recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.example-canvas__stage',
    context: 'Isolated preview stage surface',
    destination: 'Paper Surface without its own enclosing border',
  },
  {
    key: 'src/styles/shell.css:surface-style:.mobile-nav-backdrop',
    context: 'Responsive catalogue drawer scrim',
    destination: 'Paper drawer primitive built on the public overlay API',
  },
  {
    key: 'src/styles/shell.css:surface-style:.mobile-nav-drawer',
    context: 'Responsive catalogue drawer surface',
    destination: 'Paper drawer primitive built on the public overlay API',
  },
  {
    key: 'src/styles/shell.css:surface-style:.mobile-nav-trigger',
    context: 'Responsive catalogue trigger surface',
    destination: 'Paper Button recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.paper-fixture-stage',
    context: 'Fixture document canvas',
    destination: 'Paper Surface without its own enclosing border',
  },
  {
    key: 'src/styles/shell.css:surface-style:.search-backdrop',
    context: 'Search dialog scrim',
    destination: 'Paper Modal',
  },
  {
    key: 'src/styles/shell.css:surface-style:.search-field input',
    context: 'Search input surface',
    destination: 'Paper Input recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.search-result:hover',
    context: 'Search result hover surface',
    destination: 'Paper command-results or navigation-item recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.shell-icon-button',
    context: 'Shell icon-only action surface',
    destination: 'Paper Button icon-only recipe',
  },
  {
    key: 'src/styles/shell.css:surface-style:.shell-search-trigger',
    context: 'Search trigger surface',
    destination: 'Paper Button recipe',
  },
]
