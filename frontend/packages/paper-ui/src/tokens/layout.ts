/**
 * Layout tokens for Tadoku's product screens: page widths, gutters and brand sizes that the applications share or
 * own. They join the theme namespaces in `theme.ts`, so screens size themselves from tokens too.
 */
import type { Responsive } from './theme'

export const layoutSpacing = {
  /** Where a same-day separator starts in an entry list: after the date gutter and its gap. */
  'entry-rule': 'calc(7rem + var(--spacing-6))',
  /** Webv3 home: the contest card's inner inset. */
  'ledger-inset': { base: '1.5rem', lg: '1.25rem', xs: '1rem' } satisfies Responsive,
  /** The application frame's side gutter and its content's block padding. */
  gutter: { base: '1.75rem', tablet: '1.25rem', sm: '1rem' } satisfies Responsive,
  'frame-top': { base: '3rem', tablet: '2rem', sm: '1.75rem' } satisfies Responsive,
  'frame-bottom': { base: '4rem', tablet: '3rem', sm: '2.5rem' } satisfies Responsive,
  /** The header's wordmark; the mark and their gap scale with it. */
  brand: { base: '8.875rem', tablet: '7.5rem', sm: '6.625rem', xs: '5.5rem', '2xs': '4.5rem' } satisfies Responsive,
  'brand-mark': 'calc(var(--spacing-brand) * 0.282)',
  /** The Paper playground keeps its own brand steps. */
  'brand-sample': { base: '8.875rem', tablet: '7.5rem', sm: '6.625rem', '2xs': '5.5rem' } satisfies Responsive,
  'brand-sample-mark': 'calc(var(--spacing-brand-sample) * 0.282)',
  'brand-gap': 'calc(var(--spacing-brand) * 0.056)',
  'gutter-account': 'clamp(1rem, 4vw, 2rem)',
  'main-top-account': 'clamp(2rem, 7vh, 4.5rem)',
  'card-pad': 'clamp(1.5rem, 6vw, 2.25rem)',
  /** The styleguide navbar: a comfortable control plus its rule, also inside compact regions. */
  'docs-navbar': '2.8125rem',
  'docs-y': 'clamp(2rem, 5vw, 5rem)',
  'docs-x': 'clamp(1rem, 5vw, 5rem)',
  'docs-mark': '2.25rem',
  'docs-wordmark': '7.9rem',
  'docs-wordmark-y': '1.45rem',
  'motion-end': 'calc(100% - 1.5rem)',
  /** Webv3 forms and records on a surface. */
  'record-inset': 'clamp(1rem, 4vw, 2rem)',
  version: 'clamp(1rem, 4vw, 1.5rem)',
} as const

export const layoutContainer = {
  /** Webv3 home. */
  'measure-hero': { base: '17ch', tablet: '18ch' } satisfies Responsive,
  'measure-how': { base: '34ch', tablet: '54ch' } satisfies Responsive,
  'measure-message': '42ch',
  'measure-strip': '53ch',
  record: '48rem',
  catalogue: '76rem',
  'content-wide': '72rem',
  'overflow-demo': '34rem',
  'docs-section': '46rem',
  'reference-table': '32rem',
  'form-page': '42.5rem',
  'log-editor': '43.75rem',
  'log-editor-wide': '65rem',
  'editor-page': '62.75rem',
  'filter-search': '15rem',
  'filter-search-max': '27.5rem',
  filter: '11.25rem',
  'filter-narrow': '8.75rem',
  'measure-lead': '48ch',
  'measure-feature': '56ch',
  'measure-detail': '62ch',
  'measure-footnote': '76ch',
  'measure-label-sm': '10ch',
  'measure-answer': '64ch',
  'measure-context': '68ch',
  'measure-guide-title': '25ch',
  article: '45rem',
  'content-table': 'min(100%, 66ch)',
  'wide-table': '28rem',
  goal: '57.75rem',
  content: '72.5rem',
  frame: 'calc(var(--container-content) + var(--spacing-gutter) * 2)',
  card: 'min(100%, 27.5rem)',
  settings: 'min(100%, 47.5rem)',
  'settings-editor': '26rem',
  'account-label': '12rem',
} as const

export const layoutHeight = {
  'docs-sidebar': 'calc(100vh - var(--spacing-docs-navbar))',
  code: '32rem',
  'metadata-line': 'var(--leading-metadata)',
  card: '11rem',
  picker: 'min(22rem, 55vh)',
} as const

export const layoutGridTemplateColumns = {
  /** Paper styleguide. */
  'docs-shell': '18rem minmax(0, 1fr)',
  'two-one': 'minmax(0, 2fr) minmax(0, 1fr)',
  cards: 'repeat(auto-fit, minmax(min(100%, 16rem), 1fr))',
  'guide-cards': 'repeat(auto-fit, minmax(min(100%, 15rem), 1fr))',
  tiles: 'repeat(auto-fit, minmax(min(100%, 11rem), 1fr))',
  'icon-tiles': 'repeat(auto-fit, minmax(min(100%, 10rem), 1fr))',
  swatches: 'repeat(auto-fit, minmax(min(100%, 9rem), 1fr))',
  document: 'minmax(0, 48rem) 12rem',
  'document-narrow': 'minmax(0, 48rem)',
  'metadata-list': '9rem minmax(0, 1fr)',
  'space-scale': '9rem 3.5rem 1fr',
  'space-scale-narrow': '8rem 3rem 1fr',
  'log-editor': 'minmax(0, 43.75rem) minmax(13.75rem, 1fr)',
  amount: 'minmax(0, 1fr) minmax(6.875rem, 0.45fr)',
  'amount-narrow': 'minmax(0, 1fr) 6.875rem',
  'overview-columns': 'minmax(0, 1.35fr) minmax(0, 1fr)',
  'record-entry': '7rem minmax(0, 1fr) 4rem',
  'record-entry-narrow': 'minmax(0, 1fr) 3rem',
  'log-body': 'minmax(0, 1.5fr) minmax(17rem, 1fr)',
  'log-body-narrow': 'minmax(0, 1.25fr) minmax(16rem, 1fr)',
  admin: '11rem minmax(0, 1fr)',
  'admin-bottom': '1.25fr 1fr',
  'admin-record': 'minmax(0, 1fr) auto auto',
  'your-standing': 'calc(4.5rem - var(--spacing-ledger-inset) + 0.75rem) minmax(0, 1fr) auto',
  guide: '11.5rem minmax(0, 45rem)',
  'guide-narrow': '10.75rem minmax(0, 1fr)',
  timeline: '3.25rem minmax(0, 1fr)',
  /** Paper playground. */
  'main-aside': 'minmax(0, 1fr) 18.75rem',
  editor: 'minmax(0, 42.5rem) minmax(0, 16.25rem)',
  'editor-narrow': 'minmax(0, 42.5rem)',
  pair: '1fr 1fr',
  'dated-row': '3.5rem minmax(0, 1fr) auto',
  'dated-row-narrow': '2.625rem minmax(0, 1fr)',
  'standings-filters': '1fr 1fr minmax(11.25rem, 1.25fr)',
  'standings-filters-yearly': 'minmax(5.5rem, 0.6fr) 1fr 1fr minmax(11.25rem, 1.25fr)',
  /** Webv3 log form beside its suggested tags. */
  log: 'minmax(0, 40rem) minmax(12rem, 16rem)',
  /** Webv3 goal page: the form beside its summary. */
  goal: 'minmax(0, 35rem) minmax(0, 18.75rem)',
  'goal-narrow': 'minmax(0, 35rem)',
  /** A title with its status tags beside it. */
  'title-tags': 'minmax(0, max-content) auto',
  /** Webv3 standings: the board beside its aside. */
  workspace: 'minmax(0, 1fr) 16rem',
  'workspace-narrow': 'minmax(0, 1fr) 14rem',
  entry: '7rem minmax(0, 1fr) auto',
  halves: 'repeat(2, minmax(0, 1fr))',
  thirds: 'repeat(3, minmax(0, 1fr))',
  strip: '14.375rem minmax(0, 1fr) auto',
  'strip-narrow': '12.5rem minmax(0, 1fr)',
  'settings-row': '10rem minmax(0, 1fr) auto',
  'settings-row-open': '10rem minmax(0, 1fr)',
} as const

export const layoutText = {
  hero: { base: 'clamp(2rem, 3.5vw, 2.75rem)', lg: '2.25rem', tablet: 'clamp(2rem, 8.5vw, 2.75rem)' } satisfies Responsive,
  heading: '1.375rem',
  figure: '1.75rem',
  score: '1.875rem',
  measure: '2.75rem',
  avatar: '1.4rem',
  'guide-title': { base: 'clamp(2rem, 3.25vw, 2.5rem)', tablet: '2rem' } satisfies Responsive,
} as const
