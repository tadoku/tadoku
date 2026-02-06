import { DateTime } from 'luxon'
import { ContentItem } from './types'

interface Props {
  item: ContentItem
  className?: string
}

export function StatusBadge({ item, className = '' }: Props) {
  if (!item.published_at) {
    return (
      <span className={`tag bg-amber-100 text-amber-800 justify-center w-16 ${className}`}>
        Draft
      </span>
    )
  }
  const publishedAt = DateTime.fromISO(item.published_at)
  if (publishedAt > DateTime.now()) {
    return (
      <span className={`tag bg-blue-100 text-blue-800 justify-center w-20 ${className}`}>
        Scheduled
      </span>
    )
  }
  return (
    <span className={`tag bg-emerald-100 text-emerald-800 justify-center w-20 ${className}`}>
      Published
    </span>
  )
}
