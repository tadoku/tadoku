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
    key: 'src/styles/workbench.css:surface-style:.code-view pre',
    context: 'Source-code display surface',
    destination: 'Paper code-block or Surface recipe',
  },
  {
    key: 'src/styles/shell-layout.css:surface-style:.docs-header',
    context: 'Styleguide header surface',
    destination: 'Paper Navbar',
  },
  {
    key: 'src/styles/shell-layout.css:surface-style:.docs-shell',
    context: 'Styleguide application canvas',
    destination: 'Paper application-shell layout API',
  },
  {
    key: 'src/styles/shell-layout.css:surface-style:.docs-sidebar',
    context: 'Catalogue sidebar surface',
    destination: 'Paper Sidebar',
  },
  {
    key: 'src/styles/responsive.css:surface-style:.mobile-nav-drawer',
    context: 'Responsive catalogue drawer surface',
    destination: 'Paper drawer primitive built on the public overlay API',
  },
  {
    key: 'src/styles/overlays.css:surface-style:.search-result:hover',
    context: 'Search result hover surface',
    destination: 'Paper command-results or navigation-item recipe',
  },
]
