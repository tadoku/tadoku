import { useEffect } from 'react'
import { Dispatch } from 'redux'
import * as SessionStore from '../redux'
import { loadUserFromLocalStorage } from '../storage'
import { connect } from 'react-redux'

const SessionEffects = ({ loadUser }: { loadUser: () => void }) => {
  useEffect(() => loadUser(), [])

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
