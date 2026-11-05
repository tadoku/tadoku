import { useQuery } from '@tanstack/react-query'
import { Button, EmptyState, buttonClassName } from 'paper-ui'
import { Fragment, useState, type ReactNode } from 'react'
import { useLoginUrl, viewerQueryOptions, type Viewer } from '../lib/session'

/**
 * Renders a draft for the account that owns it: the first account seen signed in on this page, even if the page opened
 * signed out. The owner stays fixed when the session later ends or changes, so the draft stays visible and is never
 * saved for another account. Without an owner it explains why.
 */
export function DraftOwnerGate({ signIn, children }: {
  signIn: { title: string; description: string }
  children: (viewerId: string) => ReactNode
}) {
  const viewerQuery = useQuery(viewerQueryOptions())
  const owner = useDraftOwner(viewerQuery.data)
  const loginUrl = useLoginUrl()

  if (owner) return <Fragment key={owner}>{children(owner)}</Fragment>

  switch (viewerQuery.data?.status) {
    case 'anonymous':
      return (
        <EmptyState
          headingLevel={1}
          title={signIn.title}
          description={signIn.description}
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

function useDraftOwner(viewer: Viewer | undefined) {
  const signedIn = viewer?.status === 'signed-in' ? viewer.userId : null
  const [owner, setOwner] = useState(signedIn)
  if (!owner && signedIn) setOwner(signedIn)
  return owner ?? signedIn
}
