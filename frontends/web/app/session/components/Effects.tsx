import { useEffect } from 'react'
import { Dispatch } from 'redux'
import * as SessionStore from '../redux'
import { loadUserFromLocalStorage } from '../storage'
import { connect } from 'react-redux'
import { State } from '../../store'

const SessionEffects = ({
  loadUser,
  effectCount,
}: {
  loadUser: () => void
  effectCount: number
}) => {
  useEffect(() => loadUser(), [effectCount])

  return null
}

const mapStateToProps = (state: State) => ({
  effectCount: state.session.runEffectCount,
})

const mapDispatchToProps = (dispatch: Dispatch<SessionStore.Action>) => ({
  loadUser: () => {
    const payload = loadUserFromLocalStorage()

    if (payload) {
      dispatch({
        type: SessionStore.ActionTypes.SessionLogIn,
        payload,
      })
    }
  },
})

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SessionEffects)
