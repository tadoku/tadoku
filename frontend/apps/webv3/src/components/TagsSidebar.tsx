import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { a11y, controlStyles, withPaperForm } from 'paper-ui'
import { border, color, font, fontSize, fontWeight, glyph, lineHeight, motion, opacity, size, space } from 'paper-ui/tokens.stylex'
import { normalizeTag, parseTags, type LogDraft } from '../lib/log-draft'

const check = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3E%3Cpath d='M3 8.5l3.2 3L13 4.5' fill='none' stroke='black' stroke-width='2.5'/%3E%3C/svg%3E")`
const chevron = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3E%3Cpath d='M6 3l5 5-5 5' fill='none' stroke='black' stroke-width='2'/%3E%3C/svg%3E")`

const styles = stylex.create({
  sidebar: { gap: space[16], display: 'grid', },
  title: {
    marginInline: 0,
    color: color.textMuted,
    fontFamily: font.body,
    fontSize: fontSize[14],
    fontWeight: fontWeight.semibold,
    lineHeight: lineHeight.normal,
    marginBlockEnd: space[8],
    marginBlockStart: 0,
  },
  tags: { gap: space[8], display: 'flex', flexWrap: 'wrap', },
  tag: {
    borderColor: color.ruleFieldEdge,
    borderStyle: 'solid',
    borderWidth: border.static,
    paddingBlock: space[4],
    paddingInline: space[10],
    backgroundColor: { default: color.surfaceRaised, ':hover:not(:disabled)': color.actionSoft },
    borderBlockEndWidth: border.fieldEdge,
    boxSizing: 'border-box',
    color: color.textInk,
    cursor: { default: 'pointer', ':disabled': 'default' },
    fontSize: fontSize[13],
    minBlockSize: size[32],
    opacity: { default: null, ':disabled': opacity.inactive },
    outlineColor: { default: null, ':focus-visible': color.focusRing },
    outlineOffset: { default: null, ':focus-visible': border.focusOffset },
    outlineStyle: { default: null, ':focus-visible': 'solid' },
    outlineWidth: { default: null, ':focus-visible': border.focusRing },
  },
  tagPressed: {
    borderColor: color.textLink,
    backgroundColor: color.actionSoft,
    '::after': {
      backgroundColor: 'currentColor',
      blockSize: size[12],
      content: '""',
      display: 'inline-block',
      inlineSize: size[12],
      marginInlineStart: glyph.gap,
      maskImage: check,
      maskPosition: 'center',
      maskRepeat: 'no-repeat',
      maskSize: 'contain',
    },
  },
  summary: {
    gap: space[6],
    listStyle: 'none',
    alignItems: 'center',
    color: color.textMuted,
    cursor: 'pointer',
    display: 'inline-flex',
    fontSize: fontSize[14],
    fontWeight: fontWeight.semibold,
    marginBlockEnd: { default: null, [stylex.when.ancestor('[open]')]: space[8] },
    // Safari draws the disclosure triangle here; the rule does not list this WebKit pseudo-element.
    // eslint-disable-next-line @stylexjs/valid-styles
    '::-webkit-details-marker': { display: 'none' },
    '::before': {
      backgroundColor: 'currentColor',
      blockSize: glyph.size,
      content: '""',
      inlineSize: glyph.size,
      maskImage: chevron,
      maskPosition: 'center',
      maskRepeat: 'no-repeat',
      maskSize: 'contain',
      transform: { default: null, [stylex.when.ancestor('[open]')]: 'rotate(90deg)' },
      transitionDuration: motion.quick,
      transitionProperty: 'transform',
    },
  },
})

type SuggestedTag = { tag: string; modifier?: true }

// Webv2's per-activity lists, by activity ID. Modifier tags change the score.
const activityTags: Partial<Record<string, SuggestedTag[]>> = {
  // Reading
  1: [
    { tag: 'book' },
    { tag: 'ebook' },
    { tag: 'manga', modifier: true },
    { tag: 'comic', modifier: true },
    { tag: 'two_column', modifier: true },
    { tag: 'fiction' },
    { tag: 'non-fiction' },
    { tag: 'web page' },
    { tag: 'lyric' },
    { tag: 'game' },
    { tag: 'nsfw' },
  ],
  // Listening
  2: [
    { tag: 'dense', modifier: true },
    { tag: 'audiobook' },
    { tag: 'podcast' },
    { tag: 'anime' },
    { tag: 'drama' },
    { tag: 'tv' },
    { tag: 'news' },
    { tag: 'online video' },
    { tag: 'youtube' },
    { tag: 'fiction' },
    { tag: 'non-fiction' },
    { tag: 'nsfw' },
  ],
  // Writing
  3: [
    { tag: 'fiction' },
    { tag: 'non-fiction' },
    { tag: 'social media' },
    { tag: 'chat' },
    { tag: 'nsfw' },
  ],
  // Speaking
  4: [
    { tag: 'dense', modifier: true },
    { tag: 'conversation' },
    { tag: 'presentation' },
    { tag: 'shadowing' },
    { tag: 'chorusing' },
    { tag: 'nsfw' },
  ],
  // Study
  5: [
    { tag: 'grammar' },
    { tag: 'vocabulary' },
    { tag: 'srs' },
    { tag: 'textbook' },
    { tag: 'nsfw' },
  ],
}

export const maxTags = 10

/** The activity's score modifier and common tags, each toggling in the draft's tags until it holds `maxTags`. */
export const TagsSidebar = withPaperForm({
  defaultValues: {} as LogDraft,
  /** `compact` folds the common tags behind a disclosure, for the narrow layout where they sit inside the form. */
  props: { activityId: '', disabled: false, compact: false },
  render: function TagsSidebar({ form, activityId, disabled, compact }) {
    const tags = useStore(form.store, state => state.values.tags)
    const suggested = activityTags[activityId] ?? []
    if (!suggested.length) return null

    const selected = new Set(parseTags(tags))
    const atLimit = selected.size >= maxTags
    const toggle = (tag: string) => {
      const current = form.getFieldValue('tags')
      if (selected.has(tag)) {
        form.setFieldValue('tags', current.filter(value => normalizeTag(value) !== tag))
      } else if (!atLimit) {
        form.setFieldValue('tags', [...current, tag])
      }
    }
    const group = (title: string, items: SuggestedTag[], hideTitle = false) => items.length ? (
      <section aria-label={title}>
        <h2 {...stylex.props(styles.title, hideTitle && a11y.visuallyHidden)}>{title}</h2>
        <div {...stylex.props(styles.tags)}>
          {items.map(({ tag }) => (
            <button
              key={tag}
              type="button"
              {...stylex.props(controlStyles.inherit, styles.tag, selected.has(tag) && styles.tagPressed)}
              aria-pressed={selected.has(tag)}
              disabled={disabled || (!selected.has(tag) && atLimit)}
              onClick={() => toggle(tag)}
            >
              {tag.replaceAll('_', ' ')}
            </button>
          ))}
        </div>
      </section>
    ) : null

    const commonTags = suggested.filter(item => !item.modifier)
    const common = group('Common tags', commonTags)
    return (
      <div {...stylex.props(styles.sidebar)}>
        {group('Score modifiers', suggested.filter(item => item.modifier))}
        {compact && common ? (
          <details {...stylex.props(stylex.defaultMarker())}>
            <summary {...stylex.props(styles.summary)}>Common tags</summary>
            {group('Common tags', commonTags, true)}
          </details>
        ) : common}
      </div>
    )
  },
})
