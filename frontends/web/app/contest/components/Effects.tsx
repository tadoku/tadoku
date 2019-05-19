import { useEffect } from 'react'
import { Dispatch } from 'redux'
import * as ContestStore from '../redux'
import { connect } from 'react-redux'
import { Contest } from '../interfaces'
import ContestApi from '../api'

interface Props {
  updateLatestContest: (contest: Contest | undefined) => void
}

const ContestEffects = ({ updateLatestContest }: Props) => {
  useEffect(() => {
    const update = async () => updateLatestContest(await ContestApi.getLatest())

    update()
  }, [])

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
