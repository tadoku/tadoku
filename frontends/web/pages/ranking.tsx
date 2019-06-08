import { connect } from 'react-redux'
import { State } from '../app/store'
import RankingOverview from '../app/ranking/pages/RankingOverview'
import { Dispatch } from 'redux'
import * as RankingStore from '../app/ranking/redux'

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

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(RankingOverview)
