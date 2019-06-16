import React from 'react'
import ErrorPage from 'next/error'
import Layout from '../../ui/components/Layout'
import { SettingsTab } from '../interfaces'
import { connect } from 'react-redux'
import { State } from '../../store'
import { User } from '../../session/interfaces'

interface Props {
  tab: SettingsTab
  user: User | undefined
  userLoaded: boolean
}

const Settings = ({ tab, user, userLoaded }: Props) => {
  if (!userLoaded) {
    return <Layout title="Settings">Loading...</Layout>
  }

  if (userLoaded && !user) {
    return <ErrorPage statusCode={404} />
  }

  return <Layout title="Settings">Active tab: {tab}</Layout>
}

const mapStateToProps = (state: State) => ({
  user: state.session.user,
  userLoaded: state.session.loaded,
})

export default connect(mapStateToProps)(Settings)
