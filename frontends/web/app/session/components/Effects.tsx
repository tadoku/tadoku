import { useEffect } from 'react'
import { Dispatch } from 'redux'
import * as SessionStore from '../redux'
import { loadUserFromLocalStorage } from '../storage'
import { connect, useSelector } from 'react-redux'
import { State } from '../../store'

const SessionEffects = ({ loadUser }: { loadUser: () => void }) => {
  const effectCount = useSelector(
    (state: State) => state.session.runEffectCount,
  )
  useEffect(() => loadUser(), [effectCount])

  return null
}

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
  null,
  mapDispatchToProps,
)(SessionEffects)
