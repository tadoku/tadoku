import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { Button, EmptyState, buttonClassName } from 'paper-ui'
import { useState } from 'react'
import { LogForm } from '../../components/LogForm'
import { logOptionsQuery } from '../../lib/logs'
import { useLoginUrl, viewerQueryOptions, type Viewer } from '../../lib/session'

export const Route = createFileRoute('/logs/new')({
  head: () => ({ meta: [{ title: 'Log activity · Tadoku' }] }),
  loader: async ({ context }) => {
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    if (viewer.status === 'signed-in') await context.queryClient.ensureQueryData(logOptionsQuery(viewer.userId))
  },
  component: NewLog,
})

function NewLog() {
  const viewerQuery = useQuery(viewerQueryOptions())
  const draftOwner = useDraftOwner(viewerQuery.data)
  const loginUrl = useLoginUrl()

  if (draftOwner) return <DraftForm key={draftOwner} viewerId={draftOwner} />

  switch (viewerQuery.data?.status) {
    case 'anonymous':
      return (
        <EmptyState
          headingLevel={1}
          title="Sign in to log activity"
          description="Log in to keep your learning record and submit activity to your contests."
          actions={<a className={buttonClassName()} href={loginUrl}>Log in</a>}
        />
      )
    case 'banned':
      return (
        <EmptyState
          headingLevel={1}
          title="Your account is suspended"
          description="Activity cannot be saved for this account."
        />
      )
    default:
      return (
        <EmptyState
          headingLevel={1}
          title="Account service unavailable"
          description="We could not check your session. Try again."
          actions={<Button onClick={() => void viewerQuery.refetch()}>Try again</Button>}
        />
      )
  }
}

/**
 * The account that owns this page's draft: the first account seen signed in here, even if the page opened signed out.
 * It stays fixed when the session later ends or changes, so the draft stays visible and is never saved for another
 * account.
 */
function useDraftOwner(viewer: Viewer | undefined) {
  const signedIn = viewer?.status === 'signed-in' ? viewer.userId : null
  const [owner, setOwner] = useState(signedIn)
  if (!owner && signedIn) setOwner(signedIn)
  return owner ?? signedIn
}

function DraftForm({ viewerId }: { viewerId: string }) {
  const optionsQuery = useQuery(logOptionsQuery(viewerId))
  const loginUrl = useLoginUrl()
  const result = optionsQuery.data

  if (result?.ok) return <LogForm options={result.data} viewerId={viewerId} />
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
