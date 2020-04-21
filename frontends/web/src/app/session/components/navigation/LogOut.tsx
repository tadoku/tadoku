import React from 'react'
import Router from 'next/router'
import { ButtonLink } from '../../../ui/components'
import { connect } from 'react-redux'
import { logOut } from '../../redux'
import { removeUserFromLocalStorage } from '../../storage'
import { Dispatch } from '../../../store'

interface Props {
  logOut: () => void
}

export const LogOut = ({ logOut }: Props) => {
  return (
    <ButtonLink
      plain
      icon="sign-out-alt"
      onClick={() => {
        logOut()
        Router.push('/')
      }}
    >
      Log out
    </ButtonLink>
  )
}

const mapDispatchToProps = (dispatch: Dispatch) => ({
  logOut: () => {
    removeUserFromLocalStorage()
    dispatch(logOut())
  },
})

export default connect(null, mapDispatchToProps)(LogOut)
