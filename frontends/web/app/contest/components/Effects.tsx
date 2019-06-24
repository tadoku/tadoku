import * as ContestStore from '../redux'
import { useDispatch } from 'react-redux'
import { Contest } from '../interfaces'
import ContestApi from '../api'
import { useCachedApiState } from '../../cache'

const ContestEffects = () => {
  const dispatch = useDispatch()

  const updateLatestContest = (contest: Contest | undefined) => {
    dispatch({
      type: ContestStore.ActionTypes.ContestUpdateLatestContest,
      payload: {
        latestContest: contest,
      },
    })
  }

  useCachedApiState({
    cacheKey: `latest_contest?i=1`,
    defaultValue: undefined as Contest | undefined,
    fetchData: ContestApi.getLatest,
    onChange: updateLatestContest,
  })

  return null
}

export default ContestEffects
