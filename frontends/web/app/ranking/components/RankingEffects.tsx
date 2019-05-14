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
}

const RankingEffects = ({ user, updateRegistration }: Props) => {
  useEffect(() => {
    const update = async () => {
      if (!user) {
        updateRegistration(undefined)
        return
      }

      updateRegistration(await RankingApi.getCurrentRegistration())
    }

    update()
  }, [user])

  return null
}

const mapStateToProps = (state: State) => ({
  user: state.session.user,
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
