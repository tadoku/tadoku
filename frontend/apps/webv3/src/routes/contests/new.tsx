import * as stylex from '@stylexjs/stylex'
import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { Button, buttonStyles, EmptyState } from 'paper-ui'
import { ContestForm } from '../../components/ContestForm'
import { DraftOwnerGate } from '../../components/DraftOwnerGate'
import { contestEditorQuery, contestsHref } from '../../lib/contests'
import { useLoginUrl, viewerQueryOptions } from '../../lib/session'

export const Route = createFileRoute('/contests/new')({
  head: () => ({ meta: [{ title: 'Create new contest · Tadoku' }] }),
  loader: async ({ context }) => {
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    if (viewer.status === 'signed-in') await context.queryClient.ensureQueryData(contestEditorQuery(viewer.userId))
  },
  component: NewContest,
})

const signIn = {
  title: 'Sign in to create a contest',
  description: 'Log in to organize a community contest. You can still browse official rounds and community contests.',
}

function NewContest() {
  return (
    <DraftOwnerGate signIn={signIn} suspended="Contests cannot be created for this account.">
      {viewerId => <Editor viewerId={viewerId} />}
    </DraftOwnerGate>
  )
}

/** The form when the API lets the draft's account create a contest; otherwise why not. */
function Editor({ viewerId }: { viewerId: string }) {
  const viewer = useQuery(viewerQueryOptions()).data
  const editorQuery = useQuery(contestEditorQuery(viewerId))
  const loginUrl = useLoginUrl()
  const editor = editorQuery.data

  switch (editor?.kind) {
    case 'allowed':
      return (
        <ContestForm
          options={editor.options}
          viewerId={viewerId}
          admin={viewer?.status === 'signed-in' && viewer.userId === viewerId && viewer.role === 'admin'}
        />
      )
    case 'forbidden':
      return (
        <EmptyState
          headingLevel={1}
          title="You can’t create a contest yet"
          description="Accounts can create contests from one month after joining, up to 12 a year."
          actions={<a {...stylex.props(...buttonStyles({ variant: 'outline' }))} href={contestsHref.official()}>Browse contests</a>}
        />
      )
    case 'signed-out':
      return (
        <EmptyState
          headingLevel={1}
          title="Your session ended"
          description="Log in again to create a contest."
          actions={<a {...stylex.props(...buttonStyles())} href={loginUrl}>Log in</a>}
        />
      )
    default:
      return (
        <EmptyState
          headingLevel={1}
          title="Contest creation unavailable"
          description="We could not check whether you can create a contest. Try again."
          actions={<Button loading={editorQuery.isFetching} onClick={() => void editorQuery.refetch()}>Try again</Button>}
        />
      )
  }
}
