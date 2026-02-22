import { useFormContext } from 'react-hook-form'

interface ActivityTagSuggestion {
  tag: string
  // Future: modifiers for score adjustments
  // modifiers?: { unitId: string; multiplier: number }[]
}

const activityTags: Record<number, ActivityTagSuggestion[]> = {
  // Reading
  1: [
    { tag: 'book' },
    { tag: 'ebook' },
    { tag: 'manga' },
    { tag: 'comic' },
    { tag: 'fiction' },
    { tag: 'non-fiction' },
    { tag: 'web page' },
    { tag: 'lyric' },
    { tag: 'game' },
  ],
  // Listening
  2: [
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
  const suggestions: ActivityTagSuggestion[] = activityId != null ? (activityTags[activityId] ?? []) : []

  if (suggestions.length === 0) return null

  const isAtLimit = tags.length >= MAX_TAGS

  const handleToggle = (tag: string) => {
    const current: string[] = getValues('tags') ?? []
    if (current.includes(tag)) {
      setValue('tags', current.filter(t => t !== tag), { shouldValidate: true })
    } else if (current.length < MAX_TAGS) {
      setValue('tags', [...current, tag], { shouldValidate: true })
    }
  }

  return (
    <div className="v-stack gap-2">
      <span className="text-sm font-medium text-slate-500">Common tags</span>
      <div className="flex flex-wrap gap-2">
        {suggestions.map(({ tag }) => {
          const isSelected = tags.includes(tag)
          return (
            <button
              key={tag}
              type="button"
              onClick={() => handleToggle(tag)}
              disabled={!isSelected && isAtLimit}
              className={`tag rounded-md cursor-pointer transition-colors ${
                isSelected
                  ? 'bg-secondary/10 border border-secondary/30 text-secondary-900'
                  : 'bg-white border border-slate-300 text-slate-700 hover:bg-slate-50'
              } ${!isSelected && isAtLimit ? 'opacity-50 cursor-not-allowed' : ''}`}
            >
              {tag}
            </button>
          )
        })}
      </div>
    </div>
  )
}
