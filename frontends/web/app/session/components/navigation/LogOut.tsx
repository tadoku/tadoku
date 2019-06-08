import React from 'react'
import Router from 'next/router'
import { NavigationBarLink } from '../../../ui/components/navigation/index'
import { connect } from 'react-redux'
import * as SessionStore from '../../redux'
import { Dispatch } from 'redux'
import { removeUserFromLocalStorage } from '../../storage'

interface Props {
  signOut: () => void
}

export const LogOut = ({ signOut }: Props) => {
  return (
    <NavigationBarLink
      href="#"
      onClick={() => {
        signOut()
        Router.push('/')
      }}
    >
      Log out
    </NavigationBarLink>
  )
}

const mapDispatchToProps = (dispatch: Dispatch<SessionStore.Action>) => ({
  signOut: () => {
    removeUserFromLocalStorage()
    dispatch({ type: SessionStore.ActionTypes.SessionSignOut })
  },
})

export default connect(
  null,
  mapDispatchToProps,
)(LogOut)
