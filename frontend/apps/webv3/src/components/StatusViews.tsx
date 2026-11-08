import { Link, useRouter } from '@tanstack/react-router'
import { Button, EmptyState, buttonClassName } from 'paper-ui'

export function NotFoundView() {
  return (
    <EmptyState
      headingLevel={1}
      title="Page not found"
      description="This address doesn’t match a page on Tadoku."
      actions={<Link to="/" className={buttonClassName()}>Return home</Link>}
    />
  )
}

export function ErrorView() {
  const router = useRouter()
  return (
    <EmptyState
      headingLevel={1}
      title="This page couldn’t load"
      description="Try loading this page again."
      actions={<Button onClick={() => void router.invalidate()}>Try again</Button>}
    />
  )
}

/** A page-shaped placeholder: the destination's title and a few rows with text-shaped blocks. */
export function PendingView() {
  return (
    <div role="status" className="grid gap-5">
      <p className="sr-only">Loading…</p>
      <div className="h-8 w-1/3 bg-rule-subtle dark:bg-rule" />
      <div className="mt-2 grid max-w-article gap-1">
        {[0, 1, 2, 3, 4].map(row => (
          <div key={row} className="grid gap-2 border-b border-rule-subtle py-4 dark:border-rule">
            <div className="h-4 w-1/2 bg-rule-subtle dark:bg-rule" />
            <div className="h-3 w-1/3 bg-rule-subtle dark:bg-rule" />
          </div>
        ))}
      </div>
    </div>
  )
}
