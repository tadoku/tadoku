import { useEffect } from 'react'
import { Dispatch } from 'redux'
import * as ContestStore from '../redux'
import { connect } from 'react-redux'
import { Contest } from '../interfaces'
import ContestApi from '../api'
import { useCachedApiState } from '../../cache'

interface Props {
  updateLatestContest: (contest: Contest | undefined) => void
}

const ContestEffects = ({ updateLatestContest }: Props) => {
  const [contest] = useCachedApiState(
    `latest_contest`,
    undefined as Contest | undefined,
    ContestApi.getLatest,
    [],
  )

  useEffect(() => {
    updateLatestContest(contest)
  }, [contest])

  return null
}

const mapDispatchToProps = (dispatch: Dispatch<ContestStore.Action>) => ({
  updateLatestContest: (contest: Contest | undefined) => {
    dispatch({
      type: ContestStore.ActionTypes.ContestUpdateLatestContest,
      payload: {
        latestContest: contest,
      },
    })
  },
})

export default connect(
  null,
  mapDispatchToProps,
)(ContestEffects)
