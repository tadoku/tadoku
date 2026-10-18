import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { Button, EmptyState, buttonClassName } from 'paper-ui'
import { LogForm } from '../../components/LogForm'
import { formQueryOptions } from '../../lib/logs'
import { useLoginUrl, viewerQueryOptions } from '../../lib/session'

export const Route = createFileRoute('/logs/new')({
  head: () => ({ meta: [{ title: 'Log activity · Tadoku' }] }),
  loader: async ({ context }) => {
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    if (viewer.status === 'signed-in') await context.queryClient.ensureQueryData(formQueryOptions(viewer.userId))
    return { viewerId: viewer.status === 'signed-in' ? viewer.userId : null }
  },
  component: NewLog,
})

function NewLog() {
  const viewerQuery = useQuery(viewerQueryOptions())
  const viewer = viewerQuery.data
  const loginUrl = useLoginUrl()
  const { viewerId } = Route.useLoaderData()
  if (viewerId && (viewer?.status === 'anonymous' || viewer?.status === 'unavailable')) return <SignedInLog key={viewerId} viewerId={viewerId} />
  if (viewer?.status === 'signed-in') return <SignedInLog key={viewer.userId} viewerId={viewer.userId} />
  if (viewer?.status === 'anonymous') return <EmptyState headingLevel={1} title="Sign in to log activity" description="Log in to keep your learning record and submit activity to your contests." actions={<a className={buttonClassName()} href={loginUrl}>Log in</a>} />
  if (viewer?.status === 'banned') return <EmptyState headingLevel={1} title="Your account is suspended" description="Activity cannot be saved for this account." />
  return <EmptyState headingLevel={1} title="Account service unavailable" description="We could not check your session. Try again." actions={<Button onClick={() => void viewerQuery.refetch()}>Try again</Button>} />
}

function SignedInLog({ viewerId }: { viewerId: string }) {
  const query = useQuery(formQueryOptions(viewerId))
  const loginUrl = useLoginUrl()
  if (!query.data?.ok) {
    const signedOut = query.data && !query.data.ok && query.data.reason === 'signed-out'
    return <EmptyState headingLevel={1} title={signedOut ? 'Your session ended' : 'Activity options unavailable'} description="Your activity has not been submitted." actions={signedOut ? <a href={loginUrl} className={buttonClassName()}>Log in</a> : <Button onClick={() => void query.refetch()}>Try again</Button>} />
  }
  return <LogForm viewerId={viewerId} {...query.data.data} />
}
