import React from 'react'
import Router from 'next/router'
import { Button } from '../../../ui/components'
import { connect } from 'react-redux'
import { logOut } from '../../redux'
import { Dispatch } from 'redux'
import { removeUserFromLocalStorage } from '../../storage'

interface Props {
  logOut: () => void
}

export const LogOut = ({ logOut }: Props) => {
  return (
    <Button
      plain
      icon="sign-out-alt"
      onClick={() => {
        logOut()
        Router.push('/')
      }}
    >
      Log out
    </Button>
  )
}

const mapDispatchToProps = (dispatch: Dispatch) => ({
  logOut: () => {
    removeUserFromLocalStorage()
    dispatch(logOut())
  },
})

export default connect(null, mapDispatchToProps)(LogOut)
