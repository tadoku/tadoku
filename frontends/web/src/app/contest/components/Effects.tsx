import { useDispatch } from 'react-redux'
import { Contest } from '../interfaces'
import ContestApi from '../api'
import { updateLatestContest } from '../redux'
import { useCachedApiState } from '../../cache'
import { ContestMapper, ContestsSerializer } from '../transform'

const ContestEffects = () => {
  const dispatch = useDispatch()

  const update = (contests: Contest[]) => {
    const rawContests = contests.map(ContestMapper.optional.toRaw)
    dispatch(updateLatestContest(rawContests[0]))
  }

  useCachedApiState({
    cacheKey: `recent_contest?i=1`,
    defaultValue: [] as Contest[],
    fetchData: async () => await ContestApi.getAll(5),
    onChange: update,
    serializer: ContestsSerializer,
  })

  return null
}

export default ContestEffects
