import { updateRegistration } from '../redux'
import { useSelector, useDispatch } from 'react-redux'
import { RankingRegistration } from '../interfaces'
import { RootState } from '../../store'
import RankingApi from '../api'
import { useCachedApiState } from '../../cache'
import { OptionalizeSerializer } from '../../transform'
import { RankingRegistrationSerializer } from '../transform'

const RankingEffects = () => {
  const user = useSelector((state: RootState) => state.session.user)
  const effectCount = useSelector(
    (state: RootState) => state.ranking.runEffectCount,
  )

  const dispatch = useDispatch()
  const update = (registration: RankingRegistration | undefined) => {
    dispatch(updateRegistration(registration))
  }

  useCachedApiState({
    cacheKey: `current_registration?i=2`,
    defaultValue: undefined as RankingRegistration | undefined,
    fetchData: () => {
      if (!user) {
        return new Promise<RankingRegistration | undefined>(resolve =>
          resolve(undefined),
        )
      }

      return RankingApi.getCurrentRegistration()
    },
    onChange: update,
    dependencies: [user, effectCount],
    serializer: OptionalizeSerializer(RankingRegistrationSerializer),
  })

  return null
}

export default RankingEffects
