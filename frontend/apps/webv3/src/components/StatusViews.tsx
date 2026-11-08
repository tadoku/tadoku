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
      description="Check your connection, then try again."
      actions={<Button onClick={() => void router.invalidate()}>Try again</Button>}
    />
  )
}

export function PendingView() {
  return (
    <div role="status" className="grid gap-4 py-2">
      <div className="h-8 w-1/3 bg-rule" />
      <div className="h-4 w-1/2 bg-rule" />
      <div className="mt-6 grid gap-3">
        {[0, 1, 2, 3, 4, 5].map(row => <div key={row} className="h-10 border-b border-rule" />)}
      </div>
      <p className="m-0 text-metadata text-ink">Loading…</p>
    </div>
  )
}
