import { tv } from './variants'

/** Vertical flow. Children lose their block margins at zero specificity, so a child's own spacing still wins. */
export const stack = tv({
  base: 'paper-stack flex min-w-0 flex-col [:where(&>*)]:my-0 [:where(&>*)]:min-w-0',
  variants: {
    gap: { 2: 'gap-2', 3: 'gap-3', 4: 'gap-4', 6: 'gap-6', 8: 'gap-8' },
  },
  defaultVariants: { gap: 4 },
})

/** A wrapping row of actions or tags with the density-aware inline gap. */
export const cluster = tv({ base: 'paper-cluster flex min-w-0 flex-wrap items-center gap-inline' })

/** A readable line length; long words break instead of overflowing. */
export const measure = tv({ base: 'paper-measure max-w-measure wrap-anywhere' })

/** Paper's leading accent rail, drawn over the host's own perimeter when it has one. */
export const accentRail = tv({
  base: [
    'relative before:pointer-events-none before:absolute before:z-lift before:w-rail before:bg-action',
    'before:inset-y-0 before:start-0 forced-colors:before:bg-forced-ink forced-colors:before:forced-color-adjust-none',
  ],
  variants: {
    bordered: { true: 'before:-inset-y-rail-outset before:-start-rail-outset' },
  },
})

/** Surface perimeter, fill and elevation. */
export const surface = tv({
  base: 'border border-rule text-ink',
  variants: {
    tone: { paper: 'bg-paper', raised: 'bg-raised', overlay: 'bg-overlay' },
    elevation: {
      flat: 'shadow-none',
      floating: 'border-rule-strong shadow-floating',
      showcase: 'border-rule-strong shadow-showcase',
    },
  },
  defaultVariants: { tone: 'paper', elevation: 'flat' },
})

/** The field edge: a soft hairline on every side. */
export const fieldEdge = tv({ base: 'rounded-none border border-field-edge' })

/** The filled action edge. */
export const actionEdge = tv({ base: 'rounded-none border border-b-edge border-button-shade' })

/** Paper's type roles. Display roles use the serif; the rest use the body face. */
export const typography = tv({
  variants: {
    role: {
      display: 'paper-type-display my-0 font-display text-display leading-display font-bold tracking-display',
      page: 'paper-type-page my-0 font-display text-page leading-page font-bold tracking-page',
      section: 'paper-type-section my-0 font-display text-section leading-section font-bold tracking-section',
      component: 'paper-type-component font-body text-component leading-component font-bold',
      body: 'paper-type-body font-body text-body leading-body font-regular',
      label: 'paper-type-label font-body text-label leading-label font-semibold tracking-label',
      metadata: 'paper-type-metadata font-body text-metadata leading-metadata font-regular text-muted',
    },
  },
})

/** The keyboard focus ring for custom interactive elements. */
export const focusRing = tv({
  base: 'focus-visible:outline-ring focus-visible:outline-focus focus-visible:outline-offset-ring',
})

/**
 * An application's frame: a full-height column with a header band, a content column within the frame width and
 * the side gutter, and a container for full-bleed bands such as the footer. Pair the header with `Navbar contained`.
 */
export const appFrame = tv({
  slots: {
    root: 'flex min-h-screen flex-col',
    header: 'border-b border-rule bg-paper',
    content: 'mx-auto w-full min-w-0 max-w-frame flex-1 px-gutter pb-frame-bottom pt-frame-top focus:outline-none',
    container: 'mx-auto max-w-content',
  },
})
