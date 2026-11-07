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
]
