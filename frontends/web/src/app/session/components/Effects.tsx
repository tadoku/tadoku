import { useEffect } from 'react'
import * as SessionStore from '../redux'
import { loadUserFromLocalStorage } from '../storage'
import { useSelector, useDispatch } from 'react-redux'
import { State } from '../../store'

const SessionEffects = () => {
  const dispatch = useDispatch()
  const effectCount = useSelector(
    (state: State) => state.session.runEffectCount,
  )

  useEffect(() => {
    const payload = loadUserFromLocalStorage()

    if (payload) {
      dispatch({
        type: SessionStore.ActionTypes.SessionLogIn,
        payload,
      })
    }
  }, [effectCount])

  return null
}

export default SessionEffects
