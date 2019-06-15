import { useEffect } from 'react'
import { Dispatch } from 'redux'
import * as RankingStore from '../redux'
import { connect } from 'react-redux'
import { RankingRegistration } from '../interfaces'
import { State } from '../../store'
import { User } from '../../session/interfaces'
import RankingApi from '../api'
import { useCachedApiState } from '../../cache'

interface Props {
  user: User | undefined
  updateRegistration: (registration: RankingRegistration | undefined) => void
  effectCount: number
}

const RankingEffects = ({ user, updateRegistration, effectCount }: Props) => {
  const { data: registration } = useCachedApiState({
    cacheKey: `current_registration`,
    defaultValue: undefined as RankingRegistration | undefined,
    fetchData: () => {
      if (!user) {
        return new Promise<RankingRegistration | undefined>(resolve =>
          resolve(undefined),
        )
      }

      return RankingApi.getCurrentRegistration()
    },
    dependencies: [user, effectCount],
  })

  useEffect(() => {
    updateRegistration(registration)
  }, [registration])

  return null
}

const mapStateToProps = (state: State) => ({
  user: state.session.user,
  effectCount: state.ranking.runEffectCount,
})

const mapDispatchToProps = (dispatch: Dispatch<RankingStore.Action>) => ({
  updateRegistration: (registration: RankingRegistration | undefined) => {
    dispatch({
      type: RankingStore.ActionTypes.RankingUpdateRegistration,
      payload: {
        registration,
      },
    })
  },
})

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(RankingEffects)
