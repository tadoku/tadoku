import { createFileRoute, notFound } from '@tanstack/react-router'
import { z } from 'zod'
import { ContestNotFound, ContestPage } from '../../../../components/ContestPage'
import { contestBoardQuery, pageNumber, standingsQuery } from '../../../../lib/leaderboard'
import { registrationQuery } from '../../../../lib/registration'
import { viewerKey, viewerQueryOptions } from '../../../../lib/session'

export const Route = createFileRoute('/contests/$id/leaderboard/$')({
  loader: async ({ context: { queryClient }, params }) => {
    if (!z.guid().safeParse(params.id).success) throw notFound()
    const viewer = await queryClient.ensureQueryData(viewerQueryOptions())
    const board = await queryClient.ensureQueryData(contestBoardQuery(params.id, viewerKey(viewer)))
    if (board.kind === 'none') throw notFound()
    if (board.kind === 'unavailable') return { title: undefined }

    await Promise.all([
      queryClient.ensureQueryData(standingsQuery({ kind: 'contest', contestId: params.id }, pageNumber(params._splat))),
      viewer.status === 'signed-in' ? queryClient.ensureQueryData(registrationQuery(params.id, viewerKey(viewer))) : null,
    ])
    return { title: board.board.contest.title }
  },
  head: ({ loaderData }) => ({ meta: [{ title: loaderData?.title ? `${loaderData.title} · Tadoku` : 'Contest · Tadoku' }] }),
  notFoundComponent: ContestNotFound,
  component: function ContestRoute() {
    const { id, _splat: page } = Route.useParams()
    return <ContestPage contestId={id} view="standings" page={pageNumber(page)} />
  },
})
