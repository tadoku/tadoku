export type CustomStyleCategory =
  | 'accessibility'
  | 'application-layout'
  | 'content'
  | 'search-result-composition'
  | 'workbench-iframe'

export interface CustomStyleAllowance {
  readonly key: string
  readonly category: CustomStyleCategory
  readonly reason: string
}

/**
 * Checked allowlist for high-risk app-owned presentation that cannot be
 * expressed by a public Paper component.
 *
 * The companion contract derives these keys from source. Paper component
 * appearance and behavior never belong here: an entry must be application
 * composition, content rendering, preview infrastructure, or accessibility.
 */
export const CUSTOM_STYLE_ALLOWANCES: readonly CustomStyleAllowance[] = [
  {
    key: 'src/documentation/ExampleCanvas.tsx:workbench-iframe#1',
    category: 'workbench-iframe',
    reason: 'A real iframe is required to exercise fixture media queries.',
  },
  {
    key: 'src/styles/base.css:custom-background:.skip-link',
    category: 'accessibility',
    reason: 'The skip link must remain visible above the sticky application shell.',
  },
  {
    key: 'src/styles/documents.css:custom-background:.foundation-motion__track span',
    category: 'content',
    reason: 'The motion foundation needs a visible token-driven marker to demonstrate duration roles.',
  },
  {
    key: 'src/styles/documents.css:custom-background:.foundation-swatch--1',
    category: 'content',
    reason: 'The color foundation must render the canvas semantic token as documentation content.',
  },
  {
    key: 'src/styles/documents.css:custom-background:.foundation-swatch--2',
    category: 'content',
    reason: 'The color foundation must render the paper semantic token as documentation content.',
  },
  {
    key: 'src/styles/documents.css:custom-background:.foundation-swatch--3',
    category: 'content',
    reason: 'The color foundation must render the ink semantic token as documentation content.',
  },
  {
    key: 'src/styles/documents.css:custom-background:.foundation-swatch--4',
    category: 'content',
    reason: 'The color foundation must render the action semantic token as documentation content.',
  },
  {
    key: 'src/styles/documents.css:custom-background:.foundation-swatch--5',
    category: 'content',
    reason: 'The color foundation must render the success semantic token as documentation content.',
  },
  {
    key: 'src/styles/documents.css:custom-background:.foundation-swatch--6',
    category: 'content',
    reason: 'The color foundation must render the first chart semantic token as documentation content.',
  },
  {
    key: 'src/styles/overlays.css:custom-background:.search-result:hover',
    category: 'search-result-composition',
    reason: 'Search results are application navigation content, not form controls.',
  },
  {
    key: 'src/styles/shell-layout.css:custom-background:.docs-sidebar',
    category: 'application-layout',
    reason: 'The desktop catalogue lane needs a distinct Paper surface from the document canvas.',
  },
  {
    key: 'src/styles/shell-layout.css:custom-background:.shell-search-trigger',
    category: 'application-layout',
    reason: 'The styleguide search utility needs a quiet raised boundary within the application navbar.',
  },
  {
    key: 'src/styles/shell-layout.css:custom-background:.shell-search-trigger:active:not(:disabled)',
    category: 'application-layout',
    reason: 'The styleguide search utility preserves its neutral navbar treatment while pressed.',
  },
  {
    key: 'src/styles/shell-layout.css:custom-background:.shell-search-trigger:hover:not(:disabled)',
    category: 'application-layout',
    reason: 'The styleguide search utility uses a token-driven neutral hover within the navbar.',
  },
  {
    key: 'src/styles/shell-layout.css:custom-background:.shell-search-trigger__shortcut',
    category: 'application-layout',
    reason: 'The slash shortcut is application chrome that needs a distinct keycap surface.',
  },
  {
    key: 'src/styles/workbench.css:custom-background:.code-view pre',
    category: 'content',
    reason: 'Source code needs a scrollable, monospace reading surface.',
  },
]
