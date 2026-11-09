import { useQuery } from '@tanstack/react-query'
import { Link, createFileRoute, notFound } from '@tanstack/react-router'
import { Button, EmptyState, buttonClassName } from 'paper-ui'
import { z } from 'zod'
import { DraftOwnerGate } from '../../components/DraftOwnerGate'
import { LogForm } from '../../components/LogForm'
import { logOptionsQuery, logQuery } from '../../lib/logs'
import { useLoginUrl, viewerQueryOptions } from '../../lib/session'

export const Route = createFileRoute('/logs/$logId_/edit')({
  head: () => ({ meta: [{ title: 'Edit log · Tadoku' }] }),
  loader: async ({ context, params }) => {
    if (!z.guid().safeParse(params.logId).success) throw notFound()
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    if (viewer.status !== 'signed-in') return

    const [log] = await Promise.all([
      context.queryClient.ensureQueryData(logQuery(params.logId, viewer.userId)),
      context.queryClient.ensureQueryData(logOptionsQuery(viewer.userId)),
    ])
    if (!log.ok && log.reason === 'not-found') throw notFound()
  },
  component: EditLog,
})

const signIn = { title: 'Sign in to edit activity', description: 'Log in to change the activity you logged.' }

function EditLog() {
  return <DraftOwnerGate signIn={signIn}>{(viewerId, admin) => <EditDraft viewerId={viewerId} admin={admin} />}</DraftOwnerGate>
}

/**
 * The log's owner or an administrator edits it; the API decides. The draft starts from the log as first read; later reads
 * never replace the form.
 */
function EditDraft({ viewerId, admin }: { viewerId: string; admin: boolean }) {
  const { logId } = Route.useParams()
  const logResult = useQuery({ ...logQuery(logId, viewerId), staleTime: Infinity })
  const optionsResult = useQuery(logOptionsQuery(viewerId))
  const loginUrl = useLoginUrl()
  const log = logResult.data
  const options = optionsResult.data

  if (logResult.isPending || optionsResult.isPending) return null
  if (log?.ok && log.data.user_id !== viewerId && !admin) {
    return (
      <EmptyState
        headingLevel={1}
        title="This is someone else's record"
        description="Only the learner who logged this activity can edit it."
        actions={<Link className="text-link hover:text-link-hover" to="/logs/$logId" params={{ logId }}>View the log</Link>}
      />
    )
  }
  if (log?.ok && options?.ok) return <LogForm options={options.data} viewerId={viewerId} target={{ kind: 'edit', log: log.data }} />

  if (log?.ok === false && log.reason === 'not-found') {
    return <EmptyState headingLevel={1} title="Activity not found" description="This log may have been deleted." />
  }
  if ((log?.ok === false && log.reason === 'signed-out') || (options?.ok === false && options.reason === 'signed-out')) {
    return (
      <EmptyState
        headingLevel={1}
        title="Your session ended"
        description="Log in again to edit this activity."
        actions={<a className={buttonClassName()} href={loginUrl}>Log in</a>}
      />
    )
  }
  return (
    <EmptyState
      headingLevel={1}
      title="Activity unavailable"
      description="We could not load this activity for editing. Try again."
      actions={<Button onClick={() => void Promise.all([logResult.refetch(), optionsResult.refetch()])}>Try again</Button>}
    />
  )
}
