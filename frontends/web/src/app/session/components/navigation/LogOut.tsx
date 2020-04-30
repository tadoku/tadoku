import React from 'react'
import Router from 'next/router'
import { useDispatch } from 'react-redux'

import { Button } from '../../../ui/components'
import { logOut } from '../../redux'

export const LogOut = () => {
  const dispatch = useDispatch()
  const logOutHandler = () => {
    dispatch(logOut())
    Router.push('/')
  }

  return (
    <Button plain icon="sign-out-alt" onClick={logOutHandler}>
      Log out
    </Button>
  )
}

export default LogOut
