import { createFileRoute, notFound, redirect } from '@tanstack/react-router'
import { z } from 'zod'
import { ContestRegistrationPage } from '../../../components/ContestRegistration'
import { ContestNotFound } from '../../../components/ContestPage'
import { contestBoardQuery } from '../../../lib/leaderboard'
import { contestLanguagesQuery, registrationQuery } from '../../../lib/registration'
import { ensureViewerData, loginHref, viewerKey } from '../../../lib/session'
import { pageUnavailable } from '../../../lib/page-status'

export const Route = createFileRoute('/contests/$id/registration')({
  loader: async ({ context: { queryClient }, params, location, parentMatchPromise }) => {
    if (!z.guid().safeParse(params.id).success) throw notFound()
    const [viewer, board] = await ensureViewerData(queryClient, key => contestBoardQuery(params.id, key))
    // As in Webv2, registration needs an account: anonymous visitors log in and return here.
    if (viewer.status === 'anonymous') {
      const { accountUrl, pageUrl } = (await parentMatchPromise).loaderData!
      throw redirect({ href: loginHref(accountUrl, new URL(location.href, pageUrl).href) })
    }
    if (board.kind === 'none') throw notFound()
    if (board.kind === 'unavailable') pageUnavailable(queryClient, contestBoardQuery(params.id, viewerKey(viewer)))
    if (viewer.status !== 'signed-in') return { title: undefined }

    const { contest } = board.board
    await Promise.all([
      queryClient.ensureQueryData(registrationQuery(contest.id, viewer.userId)),
      contest.allowed_languages?.length ? null : queryClient.ensureQueryData(contestLanguagesQuery()),
    ])
    return { title: contest.title }
  },
  head: ({ loaderData }) => ({
    meta: [{ title: loaderData?.title ? `Join ${loaderData.title} · Tadoku` : 'Contest registration · Tadoku' }],
  }),
  notFoundComponent: ContestNotFound,
  component: function RegistrationRoute() {
    const { id } = Route.useParams()
    return <ContestRegistrationPage contestId={id} />
  },
})
