import { useEffect } from 'react'
import { logIn } from '../redux'
import { loadUserFromLocalStorage } from '../storage'
import { useSelector, useDispatch } from 'react-redux'
import { RootState } from '../../store'

const SessionEffects = () => {
  const dispatch = useDispatch()
  const effectCount = useSelector(
    (state: RootState) => state.session.runEffectCount,
  )

  useEffect(() => {
    const payload = loadUserFromLocalStorage()

    if (payload) {
      dispatch(logIn(payload))
    }
  }, [effectCount])

  return null
}

export default SessionEffects
