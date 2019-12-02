import React from 'react'
import Router from 'next/router'
import { Button } from '../../../ui/components'
import { connect } from 'react-redux'
import * as SessionStore from '../../redux'
import { Dispatch } from 'redux'
import { removeUserFromLocalStorage } from '../../storage'

interface Props {
  signOut: () => void
}

export const LogOut = ({ signOut }: Props) => {
  return (
    <Button
      plain
      icon="sign-out-alt"
      onClick={() => {
        signOut()
        Router.push('/')
      }}
    >
      Log out
    </Button>
  )
}

const mapDispatchToProps = (dispatch: Dispatch<SessionStore.Action>) => ({
  signOut: () => {
    removeUserFromLocalStorage()
    dispatch({ type: SessionStore.ActionTypes.SessionSignOut })
  },
})

export default connect(null, mapDispatchToProps)(LogOut)
