import { useDispatch } from 'react-redux'
import { Contest } from '../interfaces'
import ContestApi from '../api'
import { updateLatestContest } from '../redux'
import { useCachedApiState } from '../../cache'
import { ContestMapper, ContestSerializer } from '../transform'
import { OptionalizeSerializer } from '../../transform'

const ContestEffects = () => {
  const dispatch = useDispatch()

  const update = (contest: Contest | undefined) => {
    const payload = ContestMapper.optional.toRaw(contest)
    dispatch(updateLatestContest(payload))
  }

  useCachedApiState({
    cacheKey: `latest_contest?i=2`,
    defaultValue: undefined as Contest | undefined,
    fetchData: ContestApi.getLatest,
    onChange: update,
    serializer: OptionalizeSerializer(ContestSerializer),
  })

  return null
}

export default ContestEffects
