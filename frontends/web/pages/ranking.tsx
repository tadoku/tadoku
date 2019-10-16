import React from 'react'
import Head from 'next/head'
import { connect } from 'react-redux'
import { State } from '../src/store'
import RankingOverview from '../src/ranking/pages/RankingOverview'
import { Dispatch } from 'redux'
import * as RankingStore from '../src/ranking/redux'
import { Contest } from '../src/contest/interfaces'
import { RankingRegistration } from '../src/ranking/interfaces'
import { User } from '../src/session/interfaces'

const mapStateToProps = (state: State) => ({
  contest: state.contest.latestContest,
  registration: state.ranking.registration,
  user: state.session.user,
  effectCount: state.ranking.runEffectCount,
})

const mapDispatchToProps = (dispatch: Dispatch<RankingStore.Action>) => ({
  refreshRegistration: () => {
    dispatch({
      type: RankingStore.ActionTypes.RankingRunEffects,
    })
  },
})

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
