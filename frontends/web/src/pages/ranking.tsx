import React from 'react'
import Head from 'next/head'
import { connect } from 'react-redux'
import { RootState } from '../app/store'
import RankingOverview from '@app/ranking/pages/RankingOverview'
import { runEffects } from '@app/ranking/redux'
import { RawContest } from '../app/contest/interfaces'
import { RankingRegistration } from '@app/ranking/interfaces'
import { User } from '@app/session/interfaces'
import { contestMapper } from '../app/contest/transform'
import { rankingRegistrationMapper } from '@app/ranking/transform/ranking-registration'

const mapStateToProps = (state: RootState) => ({
  rawContest: state.contest.latestContest,
  registration: rankingRegistrationMapper.optional.fromRaw(
    state.ranking.rawRegistration,
  ),
  user: state.session.user,
  effectCount: state.ranking.runEffectCount,
})

const mapDispatchToProps = {
  refreshRanking: runEffects,
}

interface Props {
  rawContest: RawContest | undefined
  registration: RankingRegistration | undefined
  user: User | undefined
  effectCount: number
  refreshRanking: () => void
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(({ rawContest, ...props }: Props) => {
  if (!rawContest) {
    return null
  }

  const contest = contestMapper.fromRaw(rawContest)

  return (
    <>
      <Head>
        <title>Tadoku - Ranking</title>
      </Head>
      <RankingOverview contest={contest} {...props} />
    </>
  )
})
