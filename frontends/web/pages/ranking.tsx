import { connect } from 'react-redux'
import { State } from '../app/store'
import RankingOverview from '../app/ranking/components/RankingOverview'

const mapStateToProps = (state: State) => ({
  contest: state.contest.latestContest,
  registration: state.ranking.registration,
  user: state.session.user,
})

export default connect(mapStateToProps)(RankingOverview)
