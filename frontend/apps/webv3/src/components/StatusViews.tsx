import { Link, useRouter } from '@tanstack/react-router'
import { Button, EmptyState, Loading, PageHeader } from 'paper-ui'

export function NotFoundView() {
  return (
    <>
      <PageHeader title="Page not found" />
      <Link to="/" className="text-link hover:text-link-hover">Return home</Link>
    </>
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
    <div className="flex items-center gap-3 py-16 text-muted" role="status">
      <Loading size="small" />
      <span>Loading…</span>
    </div>
  )
}
