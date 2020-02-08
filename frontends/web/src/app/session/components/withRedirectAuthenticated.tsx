import React from 'react'
import { connect } from 'react-redux'
import { RootState } from '../../store'
import { User } from '../interfaces'
import Router from 'next/router'

export const withRedirectAuthenticated = (Component: () => JSX.Element) =>
  connect((state: RootState) => ({
    user: state.session.user,
  }))(({ user }: { user: User | undefined }) => {
    if (user) {
      Router.push('/')
      return null
    }

    return <Component />
  })
