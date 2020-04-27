import { useDispatch } from 'react-redux'
import { Contest } from '../interfaces'
import ContestApi from '../api'
import { updateRecentContests } from '../redux'
import { useCachedApiState } from '../../cache'
import { contestMapper, contestCollectionSerializer } from '../transform'

const ContestEffects = () => {
  const dispatch = useDispatch()

  const update = (contests: Contest[]) => {
    const rawContests = contests.map(contestMapper.toRaw)
    dispatch(updateRecentContests(rawContests))
  }

  useCachedApiState({
    cacheKey: `recent_contest?i=1`,
    defaultValue: [] as Contest[],
    fetchData: async () => await ContestApi.getAll(5),
    onChange: update,
    serializer: contestCollectionSerializer,
  })

  return null
}

export default ContestEffects
