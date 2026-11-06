import { useStore } from '@tanstack/react-form'
import { withPaperForm } from 'paper-ui'
import { normalizeTag, parseTags, type LogDraft } from '../lib/log-draft'

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
    const group = (title: string, items: SuggestedTag[]) => items.length ? (
      <section className="log-tags-sidebar__group" aria-label={title}>
        <h2>{title}</h2>
        <div className="log-tags-sidebar__tags">
          {items.map(({ tag }) => (
            <button
              key={tag}
              type="button"
              className="log-tags-sidebar__tag"
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

    const common = group('Common tags', suggested.filter(item => !item.modifier))
    return (
      <div className="log-tags-sidebar">
        {group('Score modifiers', suggested.filter(item => item.modifier))}
        {compact && common ? (
          <details className="log-tags-sidebar__more">
            <summary>Common tags</summary>
            {common}
          </details>
        ) : common}
      </div>
    )
  },
})
