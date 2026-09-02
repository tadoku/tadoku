import React from 'react'
import { useFormContext } from 'react-hook-form'

interface ActivityTagSuggestion {
  tag: string
  modifier?: boolean
}

const activityTags: Record<number, ActivityTagSuggestion[]> = {
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
    { tag: 'fiction' },
    { tag: 'non-fiction' },
  ],
  // Writing
  3: [
    { tag: 'fiction' },
    { tag: 'non-fiction' },
    { tag: 'social media' },
    { tag: 'chat' },
  ],
  // Speaking
  4: [
    { tag: 'dense', modifier: true },
    { tag: 'conversation' },
    { tag: 'presentation' },
    { tag: 'shadowing' },
    { tag: 'chorusing' },
  ],
  // Study
  5: [
    { tag: 'grammar' },
    { tag: 'vocabulary' },
    { tag: 'srs' },
    { tag: 'textbook' },
  ],
}

const MAX_TAGS = 10

interface TagsSidebarProps {
  activityId: number | undefined
}

export function TagsSidebar({ activityId }: TagsSidebarProps) {
  const { watch, getValues, setValue } = useFormContext()
  const tags: string[] = watch('tags') ?? []
  const allSuggestions = activityId != null ? (activityTags[activityId] ?? []) : []
  const modifiers = allSuggestions.filter(({ modifier }) => modifier)
  const suggestions = allSuggestions.filter(({ modifier }) => !modifier)

  if (modifiers.length === 0 && suggestions.length === 0) return null

  const isAtLimit = tags.length >= MAX_TAGS

  const handleToggle = (tag: string) => {
    const current: string[] = getValues('tags') ?? []
    if (current.includes(tag)) {
      setValue('tags', current.filter(t => t !== tag), { shouldValidate: true })
    } else if (current.length < MAX_TAGS) {
      setValue('tags', [...current, tag], { shouldValidate: true })
    }
  }

  const renderSuggestions = (items: ActivityTagSuggestion[]) => (
    <div className="flex flex-wrap gap-2">
      {items.map(({ tag }) => {
        const isSelected = tags.includes(tag)
        return (
          <button
            key={tag}
            type="button"
            onClick={() => handleToggle(tag)}
            disabled={!isSelected && isAtLimit}
            className={`tag cursor-pointer transition-colors border border-b-2 border-black/5 lg:border-black/15 bg-black/5 lg:bg-white ${
              isSelected ? 'opacity-70 line-through' : 'hover:bg-slate-50'
            } ${
              !isSelected && isAtLimit
                ? 'opacity-50 cursor-not-allowed'
                : ''
            }`}
          >
            {tag}
          </button>
        )
      })}
    </div>
  )

  return (
    <div className="v-stack gap-4 mt-4 lg:mt-0">
      {modifiers.length > 0 && (
        <section className="v-stack gap-2">
          <span className="subtitle text-sm">Score modifiers</span>
          {renderSuggestions(modifiers)}
        </section>
      )}
      {suggestions.length > 0 && (
        <section className="v-stack gap-2">
          <span className="subtitle text-sm">Common tags</span>
          {renderSuggestions(suggestions)}
        </section>
      )}
    </div>
  )
}
