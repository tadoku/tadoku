import React from 'react'
import { connect } from 'react-redux'
import { State } from '../../../store'
import { User } from '../../user/User'
import Router from 'next/router'

export const withRedirectAuthenticated = (Component: () => JSX.Element) =>
  connect((state: State) => ({
    user: state.user,
  }))(({ user }: { user: User | undefined }) => {
    if (user) {
      Router.push('/')
      return null
    }

    return <Component />
  })
