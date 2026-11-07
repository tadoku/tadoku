import * as stylex from '@stylexjs/stylex'
import { Link, useRouter } from '@tanstack/react-router'
import { Button, EmptyState, Loading, PageHeader, linkStyles } from 'paper-ui'
import { space } from 'paper-ui/tokens.stylex'

const styles = stylex.create({ pending: { paddingBlock: space[64] } })

export function NotFoundView() {
  return (
    <>
      <PageHeader title="Page not found" />
      <Link to="/" {...stylex.props(linkStyles.base)}>Return home</Link>
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
      actions={<Button variant="outline" onClick={() => void router.invalidate()}>Try again</Button>}
    />
  )
}

export function PendingView() {
  return <Loading xstyle={styles.pending} label="Loading page" />
}
