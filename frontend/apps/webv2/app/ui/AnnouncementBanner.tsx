import { useActiveAnnouncements } from '@app/content/api'
import { Flash } from 'ui/components/Flash'
import { ArrowTopRightOnSquareIcon, XMarkIcon } from '@heroicons/react/20/solid'
import ReactMarkdown from 'react-markdown'
import rehypeSanitize from 'rehype-sanitize'
import remarkGfm from 'remark-gfm'
import { useCallback, useEffect, useState } from 'react'
import { useRouter } from 'next/router'

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

function getPathname(href: string): string {
  try {
    return new URL(href, window.location.origin).pathname
  } catch {
    return href
  }
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
  const { asPath } = useRouter()
  const [dismissedIds, setDismissedIds] = useState<string[]>([])

  useEffect(() => {
    if (!announcements) return
    const activeIds = new Set(announcements.map(a => a.id))
    const dismissed = getDismissedIds().filter(id => activeIds.has(id))
    setDismissedIds(dismissed)
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(dismissed))
    } catch {}
  }, [announcements])

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
    <div className="relative z-10 mb-4 space-y-2">
      {visible.map(announcement => (
        <div key={announcement.id} className="relative">
          <Flash
            style={announcement.style}
            href={announcement.href && getPathname(announcement.href) !== asPath ? announcement.href : undefined}
            className="pr-10"
          >
            <div className="auto-format text-sm [&_p]:m-0 [&_p+p]:mt-1">
              <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]}>
                {announcement.content}
              </ReactMarkdown>
              {announcement.href && getPathname(announcement.href) !== asPath ? <span className="inline-flex items-center gap-1 mt-1 underline">Learn more <ArrowTopRightOnSquareIcon className="w-3.5 h-3.5" /></span> : null}
            </div>
          </Flash>
          <button
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
              handleDismiss(announcement.id)
            }}
            className="absolute right-3 top-2 p-1 hover:backdrop-brightness-90 hover:backdrop-saturate-150 transition-colors"
            title="Dismiss"
            aria-label="Dismiss announcement"
          >
            <XMarkIcon className="w-4 h-4" />
          </button>
        </div>
      ))}
    </div>
  )
}
