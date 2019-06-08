import { ExpressNextContext } from '../app/interfaces'
import RankingProfile from '../app/ranking/components/RankingProfile'
import { connect } from 'react-redux'
import { State } from '../app/store'
import { Dispatch } from 'redux'
import * as RankingStore from '../app/ranking/redux'

interface Props {
  contestId: number | undefined
  userId: number | undefined
  effectCount: number
  refreshRanking: () => void
}

const RankingDetails = (props: Props) => <RankingProfile {...props} />

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
