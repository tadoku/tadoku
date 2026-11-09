import { createFileRoute, notFound } from '@tanstack/react-router'
import { z } from 'zod'
import { ContestNotFound, ContestPage } from '../../../../components/ContestPage'
import { contestLogsQuery } from '../../../../lib/contest-activity'
import { contestBoardQuery, pageNumber } from '../../../../lib/leaderboard'
import { registrationQuery } from '../../../../lib/registration'
import { ensureViewerData, viewerKey } from '../../../../lib/session'
import { pageUnavailable } from '../../../../lib/page-status'

export const Route = createFileRoute('/contests/$id/updates/$')({
  loader: async ({ context: { queryClient }, params }) => {
    if (!z.guid().safeParse(params.id).success) throw notFound()
    const boardRead = ensureViewerData(queryClient, key => contestBoardQuery(params.id, key))
    const [[viewer, board]] = await Promise.all([
      boardRead,
      queryClient.ensureQueryData(contestLogsQuery({ contestId: params.id, page: pageNumber(params._splat) })),
      boardRead.then(([viewer]) => (viewer.status === 'signed-in'
        ? queryClient.ensureQueryData(registrationQuery(params.id, viewerKey(viewer)))
        : null)),
    ])
    if (board.kind === 'none') throw notFound()
    if (board.kind === 'unavailable') pageUnavailable(queryClient, contestBoardQuery(params.id, viewerKey(viewer)))
    return { title: board.kind === 'board' ? board.board.contest.title : undefined }
  },
  head: ({ loaderData }) => ({
    meta: [{ title: loaderData?.title ? `Activity · ${loaderData.title} · Tadoku` : 'Contest activity · Tadoku' }],
  }),
  notFoundComponent: ContestNotFound,
  component: function ContestUpdatesRoute() {
    const { id, _splat: page } = Route.useParams()
    return <ContestPage contestId={id} view="updates" page={pageNumber(page)} />
  },
})
