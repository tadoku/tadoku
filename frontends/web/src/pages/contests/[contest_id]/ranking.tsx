import React from 'react'
import Head from 'next/head'
import ErrorPage from 'next/error'
import { useRouter } from 'next/router'
import { useSelector, useDispatch } from 'react-redux'

import { RootState } from '../../../app/store'
import RankingOverview from '../../../app/ranking/pages/RankingOverview'
import { runEffects } from '../../../app/ranking/redux'
import { ContestSerializer } from '../../../app/contest/transform'
import { RankingRegistrationMapper } from '../../../app/ranking/transform/ranking-registration'
import { useCachedApiState, isReady } from '../../../app/cache'
import ContestApi from '../../../app/contest/api'
import { OptionalizeSerializer } from '../../../app/transform'
import { Contest } from '../../../app/contest/interfaces'

export default () => {
  const registration = useSelector((state: RootState) =>
    RankingRegistrationMapper.optional.fromRaw(state.ranking.rawRegistration),
  )
  const user = useSelector((state: RootState) => state.session.user)
  const effectCount = useSelector(
    (state: RootState) => state.ranking.runEffectCount,
  )
  const dispatch = useDispatch()
  const refreshRanking = () => dispatch(runEffects())

  const router = useRouter()
  const contestId = parseInt(router.query.contest_id as string)

  // TODO: extract out cache keys in a central place
  const { data: contest, status: statusContest } = useCachedApiState<
    Contest | undefined
  >({
    cacheKey: `contest?i=1&id=${contestId}`,
    defaultValue: undefined,
    fetchData: () => {
      return ContestApi.get(contestId)
    },
    dependencies: [contestId],
    serializer: OptionalizeSerializer(ContestSerializer),
  })

  if (!contestId) {
    return <ErrorPage statusCode={404} />
  }

  if (!isReady([statusContest])) {
    return <p>Loading...</p>
  }

  if (!contest) {
    return <ErrorPage statusCode={404} />
  }

  return (
    <>
      <Head>
        <title>Tadoku - Ranking for {contest.description}</title>
      </Head>
      <RankingOverview
        contest={contest}
        registration={registration}
        user={user}
        effectCount={effectCount}
        refreshRanking={refreshRanking}
      />
    </>
  )
}
