import { useQuery } from '@tanstack/react-query'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { Button, EmptyState, buttonClassName } from 'paper-ui'
import { z } from 'zod'
import { ContestScoring } from '../../../components/ContestScoring'
import { getPageUrls } from '../../../lib/links'
import { contestScoringQuery } from '../../../lib/scoring'
import { useLoginUrl, viewerQueryOptions } from '../../../lib/session'

const isContestId = (id: string) => z.guid().safeParse(id).success

/** Contest scoring is an administrator tool, as in Webv2. Anonymous visitors are sent to log in. */
export const Route = createFileRoute('/contests/$id/scoring')({
  head: () => ({ meta: [{ title: 'Contest scoring · Tadoku' }] }),
  loader: async ({ context, params, location }) => {
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    if (viewer.status === 'anonymous') {
      const { accountUrl, pageUrl } = await getPageUrls()
      const returnTo = new URL(location.href, pageUrl).href
      throw redirect({ href: `${accountUrl}/login?return_to=${encodeURIComponent(returnTo)}`, reloadDocument: true })
    }
    if (viewer.status === 'signed-in' && viewer.role === 'admin' && isContestId(params.id)) {
      await context.queryClient.ensureQueryData(contestScoringQuery(params.id, viewer.userId))
    }
  },
  component: Scoring,
})

function Scoring() {
  const { id } = Route.useParams()
  const viewerQuery = useQuery(viewerQueryOptions())
  const loginUrl = useLoginUrl()
  const viewer = viewerQuery.data

  if (viewer?.status === 'signed-in' && viewer.role === 'admin') {
    return isContestId(id)
      ? <ContestScoring key={viewer.userId} contestId={id} viewerId={viewer.userId} />
      : <EmptyState headingLevel={1} title="Contest not found" description="This contest does not exist." />
  }
  switch (viewer?.status) {
    case 'signed-in':
    case 'banned':
      return <EmptyState headingLevel={1} title="This feature is not yet available." />
    case 'anonymous':
      return (
        <EmptyState
          headingLevel={1}
          title="Sign in to manage contest scoring"
          actions={<a className={buttonClassName()} href={loginUrl}>Log in</a>}
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
