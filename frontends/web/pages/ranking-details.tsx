import React from 'react'
import Head from 'next/head'
import ErrorPage from 'next/error'
import { ExpressNextContext } from '../app/interfaces'
import RankingProfile from '../app/ranking/pages/RankingProfile'
import { connect } from 'react-redux'
import { State } from '../app/store'
import { Dispatch } from 'redux'
import * as RankingStore from '../app/ranking/redux'

interface Props {
  contestId: number | undefined
  userId: number | undefined
  effectCount: number
  refreshRanking: () => void
  dispatch: Dispatch
}

const RankingDetails = ({ contestId, userId, ...props }: Props) => {
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

RankingDetails.getInitialProps = async ({ req, query }: ExpressNextContext) => {
  if (req && req.params) {
    const { contest_id, user_id } = req.params

    return {
      contestId: parseInt(contest_id),
      userId: parseInt(user_id),
    }
  }

  if (query.contest_id && query.user_id) {
    const { contest_id, user_id } = query

    return {
      contestId: parseInt(contest_id as string),
      userId: parseInt(user_id as string),
    }
  }

  return {}
}

const mapStateToProps = (state: State) => ({
  effectCount: state.ranking.runEffectCount,
})

const mapDispatchToProps = (dispatch: Dispatch<RankingStore.Action>) => ({
  dispatch,
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
