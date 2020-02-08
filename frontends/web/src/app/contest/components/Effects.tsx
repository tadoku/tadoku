import { useDispatch } from 'react-redux'
import { Contest } from '../interfaces'
import ContestApi from '../api'
import { updateLatestContest } from '../redux'
import { useCachedApiState } from '../../cache'

const ContestEffects = () => {
  const dispatch = useDispatch()

  const update = (contest: Contest | undefined) => {
    dispatch(updateLatestContest(contest))
  }

  useCachedApiState({
    cacheKey: `latest_contest?i=1`,
    defaultValue: undefined as Contest | undefined,
    fetchData: ContestApi.getLatest,
    onChange: update,
  })

  return null
}

export default ContestEffects
