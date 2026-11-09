import { useQuery } from '@tanstack/react-query'
import { createFileRoute, Link, notFound } from '@tanstack/react-router'
import { EmptyState } from 'paper-ui'
import { z } from 'zod'
import { ProfileActivity } from '../../../../components/ProfileActivity'
import { PendingView } from '../../../../components/StatusViews'
import { pageNumber } from '../../../../lib/leaderboard'
import { profileActivityQuery } from '../../../../lib/profile'
import { viewerQueryOptions } from '../../../../lib/session'
import { pageUnavailable } from '../../../../lib/page-status'
import { sharingMeta } from '../../../../lib/page-meta'

export const Route = createFileRoute('/profile/$id/updates/{-$page}')({
  // Webv2 ignores a malformed day and lists every update.
  validateSearch: z.object({ date: z.iso.date().optional().catch(undefined) }),
  loaderDeps: ({ search }) => ({ date: search.date }),
  loader: async ({ context, params, deps }) => {
    const [, result] = await Promise.all([
      context.queryClient.ensureQueryData(viewerQueryOptions()),
      context.queryClient.ensureQueryData(profileActivityQuery(activityInput(params, deps.date))),
    ])
    if (result.status === 'not-found') throw notFound()
    if (result.status === 'unavailable') pageUnavailable(context.queryClient, profileActivityQuery(activityInput(params, deps.date)))
    const name = result.profile.display_name
    return { name, description: `${name}’s immersion logs on Tadoku.` }
  },
  head: ({ loaderData }) => ({
    meta: loaderData
      ? sharingMeta(loaderData.name, loaderData.description, `Profile updates · ${loaderData.name} · Tadoku`)
      : [{ title: 'Profile · Tadoku' }],
  }),
  notFoundComponent: () => (
    <EmptyState
      headingLevel={1}
      title="Profile not found"
      description="No Tadoku account has this address."
      actions={<Link to="/" className="text-link hover:text-link-hover">Return home</Link>}
    />
  ),
  component: ProfileActivityPage,
})

/** The requested profile, page and day; a malformed page reads as the first, as in Webv2. */
function activityInput(params: { id: string; page?: string }, date: string | undefined) {
  if (!z.guid().safeParse(params.id).success) throw notFound()
  return { userId: params.id, page: pageNumber(params.page), date }
}

function ProfileActivityPage() {
  const input = activityInput(Route.useParams(), Route.useLoaderDeps().date)
  const { data: viewer } = useQuery(viewerQueryOptions())
  const { data: result } = useQuery(profileActivityQuery(input))

  if (!result) return <PendingView />
  if (result.status !== 'found') {
    return <EmptyState headingLevel={1} title="Profile unavailable" description="Try reloading this page." />
  }
  const owner = viewer?.status === 'signed-in' && viewer.userId === result.profile.id
  return <ProfileActivity activity={result} page={input.page} date={input.date} owner={owner} />
}
