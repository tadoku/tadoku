import { useEffect } from 'react'
import { Dispatch } from 'redux'
import * as RankingStore from '../redux'
import { connect } from 'react-redux'
import { RankingRegistration } from '../interfaces'
import { State } from '../../store'
import { User } from '../../user/interfaces'
import RankingApi from '../api'

interface Props {
  user: User | undefined
  updateRegistration: (registration: RankingRegistration | undefined) => void
  effectCount: number
}

const RankingEffects = ({ user, updateRegistration, effectCount }: Props) => {
  useEffect(() => {
    const update = async () => {
      if (!user) {
        updateRegistration(undefined)
        return
      }

      updateRegistration(await RankingApi.getCurrentRegistration())
    }

    update()
  }, [user, effectCount])

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
