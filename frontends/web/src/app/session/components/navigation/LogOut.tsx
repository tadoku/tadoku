import React from 'react'
import Router from 'next/router'
import { useDispatch } from 'react-redux'

import { Button } from '../../../ui/components'

export const LogOut = () => {
  const dispatch = useDispatch()
  const logOut = () => {
    dispatch(logOut())
    Router.push('/')
  }

  return (
    <Button plain icon="sign-out-alt" onClick={logOut}>
      Log out
    </Button>
  )
}

export default LogOut
