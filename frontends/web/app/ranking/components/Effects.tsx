import { Dispatch } from 'redux'
import * as RankingStore from '../redux'
import { connect, useSelector, useDispatch } from 'react-redux'
import { RankingRegistration } from '../interfaces'
import { State } from '../../store'
import RankingApi from '../api'
import { useCachedApiState, DefaultSerializer } from '../../cache'
import { OptionalizeSerializer } from '../../transform'

const RankingEffects = () => {
  const user = useSelector((state: State) => state.session.user)
  const effectCount = useSelector(
    (state: State) => state.ranking.runEffectCount,
  )

  const dispatch = useDispatch()
  const updateRegistration = (
    registration: RankingRegistration | undefined,
  ) => {
    dispatch({
      type: RankingStore.ActionTypes.RankingUpdateRegistration,
      payload: {
        registration,
      },
    })
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
    onChange: updateRegistration,
    dependencies: [user, effectCount],
    serializer: OptionalizeSerializer(DefaultSerializer),
  })

  return null
}

export default RankingEffects
