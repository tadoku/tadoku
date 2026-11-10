import { useQuery } from '@tanstack/react-query'
import { useHydrated, useRouterState } from '@tanstack/react-router'
import { Button, Flash } from 'paper-ui'
import { XMarkIcon, iconClassName } from 'paper-ui/icons'
import { lazy, Suspense, useEffect, useState } from 'react'
import { announcementsQuery } from '../lib/content'

// Webv2's key, so dismissals carry over when the host switches.
const storageKey = 'dismissed_announcements'

function readDismissed(): string[] {
  try {
    const stored: unknown = JSON.parse(localStorage.getItem(storageKey) ?? '[]')
    return Array.isArray(stored) ? stored.filter(id => typeof id === 'string') : []
  } catch {
    return []
  }
}

function storeDismissed(ids: string[]) {
  try { localStorage.setItem(storageKey, JSON.stringify(ids)) } catch { /* Without storage a dismissal lasts for this page. */ }
}

// Announcements render only after hydration, so the Markdown renderer loads then instead of in every page's entry.
const Markdown = lazy(() => import('./Markdown'))

/**
 * Where Learn more may lead: an HTTPS page, or a path on this site. The API also accepts plain HTTP; anything else,
 * such as `//host` or a script URL, has no link.
 */
function learnMoreHref(href: string | undefined) {
  if (!href) return undefined
  if (href.startsWith('/')) return href.startsWith('//') || href.startsWith('/\\') ? undefined : href
  try {
    const url = new URL(href)
    return url.protocol === 'https:' && url.hostname && !url.username && !url.password ? href : undefined
  } catch {
    return undefined
  }
}

function pathOf(href: string) {
  try { return new URL(href, location.origin).pathname } catch { return href }
}

/** The CMS's running announcements, each until this browser dismisses it. */
export function Announcements() {
  const hydrated = useHydrated()
  const { data: announcements = [] } = useQuery({ ...announcementsQuery(), enabled: hydrated })
  const currentPath = useRouterState({ select: state => state.location.pathname })
  // The query waits for hydration, so this browser-only read never changes server-rendered output.
  const [dismissed, setDismissed] = useState(readDismissed)

  useEffect(() => {
    // Forget dismissals of announcements that have ended.
    if (announcements.length) storeDismissed(readDismissed().filter(id => announcements.some(item => item.id === id)))
  }, [announcements])

  const dismiss = (id: string) => {
    const next = [...dismissed, id]
    storeDismissed(next)
    setDismissed(next)
  }

  const visible = announcements.filter(announcement => !dismissed.includes(announcement.id))
  if (!visible.length) return null
  return (
    <div className="mb-8 grid gap-2">
      {visible.map(announcement => (
        <Flash
          key={announcement.id}
          variant={announcement.tone}
          action={
            <Button variant="ghost" aria-label="Dismiss announcement" title="Dismiss" onClick={() => dismiss(announcement.id)}>
              <XMarkIcon className={iconClassName()} aria-hidden="true" />
            </Button>
          }
        >
          <div className="grid gap-1 wrap-anywhere [&_p]:m-0">
            <Suspense fallback={null}>
              <Markdown>{announcement.content}</Markdown>
            </Suspense>
            <LearnMore href={learnMoreHref(announcement.href)} currentPath={currentPath} />
          </div>
        </Flash>
      ))}
    </div>
  )
}

function LearnMore({ href, currentPath }: { href: string | undefined; currentPath: string }) {
  if (!href || pathOf(href) === currentPath) return null
  return <a className="text-link hover:text-link-hover" href={href}>Learn more</a>
}
