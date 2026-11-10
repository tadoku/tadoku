import { useStore } from '@tanstack/react-form'
import { withPaperForm } from 'paper-ui'
import { normalizeTag, parseTags, type LogDraft } from '../lib/log-draft'
import { ChevronRightIcon } from 'paper-ui/icons'

const tagButton = [
  'min-h-8 cursor-pointer border border-field-edge bg-field px-2.5 py-1 text-fine text-ink',
  'not-disabled:hover:border-field-edge-hover not-disabled:hover:bg-neutral-hover focus-visible:outline-ring focus-visible:outline-focus',
  'focus-visible:outline-offset-ring disabled:cursor-default disabled:opacity-faded',
  // As in Webv2, a chosen tag is struck through, so choosing one never changes the chip's width.
  'aria-pressed:text-muted aria-pressed:line-through',
].join(' ')

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
        <h2 className={hideTitle ? 'sr-only' : 'mb-2 mt-0 font-body text-sm leading-150 font-semibold text-muted'}>{title}</h2>
        <div className="flex flex-wrap gap-2">
          {items.map(({ tag }) => (
            <button
              key={tag}
              type="button"
              className={tagButton}
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

    const common = suggested.filter(item => !item.modifier)
    return (
      <div className="grid gap-4">
        {group('Score modifiers', suggested.filter(item => item.modifier))}
        {compact && common.length ? (
          <details className="group/more">
            <summary
              className={[
                'inline-flex cursor-pointer list-none items-center gap-1.5 text-sm font-semibold text-muted',
                'group-open/more:mb-2 [&::-webkit-details-marker]:hidden',
              ].join(' ')}
            >
              <ChevronRightIcon
                className="size-2.5 transition-transform duration-quick group-open/more:rotate-quarter"
                aria-hidden="true"
              />
              Common tags
            </summary>
            {group('Common tags', common, true)}
          </details>
        ) : group('Common tags', common)}
      </div>
    )
  },
})
