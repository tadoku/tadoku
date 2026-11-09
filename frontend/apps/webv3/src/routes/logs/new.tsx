import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Button, EmptyState, buttonClassName } from 'paper-ui'
import { DraftOwnerGate } from '../../components/DraftOwnerGate'
import { LogForm } from '../../components/LogForm'
import { logOptionsQuery } from '../../lib/logs'
import { useLoginUrl, viewerQueryOptions } from '../../lib/session'
import { noindex } from '../../lib/page-meta'

export const Route = createFileRoute('/logs/new')({
  head: () => ({ meta: [{ title: 'Log activity · Tadoku' }, noindex] }),
  // Webv2 prefills a positive whole amount from `?amount=` and ignores anything else.
  validateSearch: z.object({ amount: z.coerce.number().int().positive().optional().catch(undefined) }),
  loader: async ({ context }) => {
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    if (viewer.status === 'signed-in') await context.queryClient.ensureQueryData(logOptionsQuery(viewer.userId))
  },
  component: NewLog,
})

const signIn = {
  title: 'Sign in to log activity',
  description: 'Log in to keep your learning record and submit activity to your contests.',
}

function NewLog() {
  return <DraftOwnerGate signIn={signIn}>{viewerId => <DraftForm viewerId={viewerId} />}</DraftOwnerGate>
}

function DraftForm({ viewerId }: { viewerId: string }) {
  const { amount } = Route.useSearch()
  const optionsQuery = useQuery(logOptionsQuery(viewerId))
  const loginUrl = useLoginUrl()
  const result = optionsQuery.data

  if (result?.ok) return <LogForm options={result.data} viewerId={viewerId} target={{ kind: 'new', amount }} />
  if (result && result.reason === 'signed-out') {
    return (
      <EmptyState
        headingLevel={1}
        title="Your session ended"
        description="Your activity has not been submitted."
        actions={<a className={buttonClassName()} href={loginUrl}>Log in</a>}
      />
    )
  }
  return (
    <EmptyState
      headingLevel={1}
      title="Activity options unavailable"
      description="Your activity has not been submitted."
      actions={<Button onClick={() => void optionsQuery.refetch()}>Try again</Button>}
    />
  )
}
