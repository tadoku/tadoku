import React from 'react'
import Head from 'next/head'
import ErrorPage from 'next/error'
import RankingProfile from '../../../ranking/pages/RankingProfile'
import { connect } from 'react-redux'
import { State } from '../../../store'
import { Dispatch } from 'redux'
import * as RankingStore from '../../../ranking/redux'
import { useRouter } from 'next/router'

interface Props {
  effectCount: number
  refreshRanking: () => void
}

const RankingDetails = (props: Props) => {
  const router = useRouter()
  const contestId = parseInt(router.query.contest_id as string)
  const userId = parseInt(router.query.user_id as string)

  if (!contestId || !userId) {
    return <ErrorPage statusCode={404} />
  }

  return (
    <>
      <Head>
        <title>Tadoku - Stats</title>
      </Head>
      <RankingProfile contestId={contestId} userId={userId} {...props} />
    </>
  )
}

const mapStateToProps = (state: State) => ({
  effectCount: state.ranking.runEffectCount,
})

const mapDispatchToProps = (dispatch: Dispatch<RankingStore.Action>) => ({
  refreshRanking: () => {
    dispatch({
      type: RankingStore.ActionTypes.RankingRunEffects,
    })
  },
})

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(RankingDetails)
