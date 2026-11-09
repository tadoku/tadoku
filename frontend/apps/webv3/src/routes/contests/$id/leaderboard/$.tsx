import { createFileRoute, notFound } from '@tanstack/react-router'
import { z } from 'zod'
import { ContestNotFound, ContestPage } from '../../../../components/ContestPage'
import { contestBoardQuery, pageNumber, standingsQuery } from '../../../../lib/leaderboard'
import { registrationQuery } from '../../../../lib/registration'
import { ensureViewerData, viewerKey } from '../../../../lib/session'
import { pageUnavailable } from '../../../../lib/page-status'
import { contestDescription, noindex, sharingMeta } from '../../../../lib/page-meta'

export const Route = createFileRoute('/contests/$id/leaderboard/$')({
  loader: async ({ context: { queryClient }, params }) => {
    if (!z.guid().safeParse(params.id).success) throw notFound()
    const boardRead = ensureViewerData(queryClient, key => contestBoardQuery(params.id, key))
    const [[viewer, board]] = await Promise.all([
      boardRead,
      queryClient.ensureQueryData(standingsQuery({ kind: 'contest', contestId: params.id }, pageNumber(params._splat))),
      boardRead.then(([viewer]) => (viewer.status === 'signed-in'
        ? queryClient.ensureQueryData(registrationQuery(params.id, viewerKey(viewer)))
        : null)),
    ])
    if (board.kind === 'none') throw notFound()
    if (board.kind === 'unavailable') pageUnavailable(queryClient, contestBoardQuery(params.id, viewerKey(viewer)))
    const { contest } = board.board
    return { title: contest.title, description: contestDescription(contest), private: contest.private }
  },
  // A private contest is open to anyone with its URL, but stays out of search results.
  head: ({ loaderData }) => ({
    meta: loaderData
      ? [...sharingMeta(loaderData.title, loaderData.description), ...(loaderData.private ? [noindex] : [])]
      : [{ title: 'Contest · Tadoku' }],
  }),
  notFoundComponent: ContestNotFound,
  component: function ContestRoute() {
    const { id, _splat: page } = Route.useParams()
    return <ContestPage contestId={id} view="standings" page={pageNumber(page)} />
  },
})
