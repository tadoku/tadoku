import React from 'react'
import Head from 'next/head'
import { connect } from 'react-redux'
import { RootState } from '../app/store'
import RankingOverview from '../app/ranking/pages/RankingOverview'
import { runEffects } from '../app/ranking/redux'
import { Contest } from '../app/contest/interfaces'
import { RankingRegistration } from '../app/ranking/interfaces'
import { User } from '../app/session/interfaces'

const mapStateToProps = (state: RootState) => ({
  contest: state.contest.latestContest,
  registration: state.ranking.registration,
  user: state.session.user,
  effectCount: state.ranking.runEffectCount,
})

const mapDispatchToProps = {
  refreshRegistration: runEffects,
}

interface Props {
  contest: Contest | undefined
  registration: RankingRegistration | undefined
  user: User | undefined
  effectCount: number
  refreshRegistration: () => void
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(({ contest, ...props }: Props) => {
  if (!contest) {
    return null
  }

  return (
    <>
      <Head>
        <title>Tadoku - Ranking</title>
      </Head>
      <RankingOverview contest={contest} {...props} />
    </>
  )
})
