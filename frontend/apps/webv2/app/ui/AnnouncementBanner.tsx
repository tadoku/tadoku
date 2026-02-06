import { useActiveAnnouncements } from '@app/content/api'
import { Flash } from 'ui/components/Flash'
import { XMarkIcon } from '@heroicons/react/20/solid'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useCallback, useEffect, useState } from 'react'

const STORAGE_KEY = 'dismissed_announcements'

function getDismissedIds(): string[] {
  if (typeof window === 'undefined') return []
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      const parsed = JSON.parse(stored)
      if (Array.isArray(parsed)) return parsed
    }
  } catch {}
  return []
}

function dismissAnnouncement(id: string) {
  const dismissed = getDismissedIds()
  if (!dismissed.includes(id)) {
    dismissed.push(id)
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(dismissed))
    } catch {}
  }
}

export default function AnnouncementBanner() {
  const { data: announcements } = useActiveAnnouncements()
  const [dismissedIds, setDismissedIds] = useState<string[]>([])

  useEffect(() => {
    setDismissedIds(getDismissedIds())
  }, [])

  const handleDismiss = useCallback((id: string) => {
    dismissAnnouncement(id)
    setDismissedIds(prev => [...prev, id])
  }, [])

  if (!announcements || announcements.length === 0) {
    return null
  }

  const visible = announcements.filter(a => !dismissedIds.includes(a.id))

  if (visible.length === 0) {
    return null
  }

  return (
    <div className="w-full">
      {visible.map(announcement => (
        <div key={announcement.id} className="relative">
          <Flash
            style={announcement.style}
            href={announcement.href ?? undefined}
            className="rounded-none border-x-0 pr-10"
          >
            <span className="auto-format text-sm [&_p]:m-0 [&_p]:inline">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {announcement.content}
              </ReactMarkdown>
            </span>
          </Flash>
          <button
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
              handleDismiss(announcement.id)
            }}
            className="absolute right-3 top-1/2 -translate-y-1/2 p-1 rounded hover:bg-black/10 transition-colors"
            aria-label="Dismiss announcement"
          >
            <XMarkIcon className="w-4 h-4" />
          </button>
        </div>
      ))}
    </div>
  )
}
