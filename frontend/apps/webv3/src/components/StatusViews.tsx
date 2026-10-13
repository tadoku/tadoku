import { Link, useRouter } from '@tanstack/react-router'
import { Button, EmptyState, Loading } from 'paper-ui'

export function NotFoundView() {
  return (
    <EmptyState
      headingLevel={1}
      title="Page not found"
      description="This address doesn’t match a Tadoku page."
      actions={<Link to="/" className="text-link">Return home</Link>}
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
      actions={<Button variant="outline" onClick={() => void router.invalidate()}>Try again</Button>}
    />
  )
}

export function PendingView() {
  return <Loading className="app-pending" label="Loading page" />
}
