import { useEffect } from 'react'
import { connect } from 'react-redux'
import { SessionActionTypes, SessionActions, State } from '../store'
import { Dispatch } from 'redux'
import { removeUserFromLocalStorage } from '../app/session/storage'
import { User } from '../app/user/User'
import Router from 'next/router'

interface Props {
  user: User | undefined
  signOut: () => void
}

const SignOut = ({ user, signOut }: Props) => {
  useEffect(() => {
    if (user) {
      signOut()
    } else {
      Router.push('/')
    }
  }, [user])

  return null
}

const mapStateToProps = (state: State) => ({
  user: state.user,
})

const mapDispatchToProps = (dispatch: Dispatch<SessionActions>) => ({
  signOut: () => {
    removeUserFromLocalStorage()
    dispatch({ type: SessionActionTypes.SessionSignOut })
  },
})

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SignOut)
