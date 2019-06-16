import React from 'react'
import ErrorPage from 'next/error'
import Layout from '../../ui/components/Layout'
import { SettingsTab } from '../interfaces'
import { connect } from 'react-redux'
import { State } from '../../store'
import { User } from '../../session/interfaces'
import SettingsSidebar from '../components/SettingsSidebar'
import styled from 'styled-components'
import ChangePasswordForm from '../components/forms/ChangePasswordForm'
import ProfileForm from '../components/forms/ProfileForm'

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

  return (
    <Layout title="Settings">
      <Container>
        <SettingsSidebar activeTab={tab} />
        <Content>{componentForTab(tab, user as User)}</Content>
      </Container>
    </Layout>
  )
}

const componentForTab = (tab: SettingsTab, user: User): JSX.Element => {
  switch (tab) {
    case SettingsTab.Profile: {
      return (
        <FormContainer>
          <h2>Profile</h2>
          <ProfileForm user={user} />
        </FormContainer>
      )
    }
    case SettingsTab.ChangePassword: {
      return (
        <FormContainer>
          <h2>Change password</h2>
          <ChangePasswordForm />
        </FormContainer>
      )
    }
  }
}

const mapStateToProps = (state: State) => ({
  user: state.session.user,
  userLoaded: state.session.loaded,
})

export default connect(mapStateToProps)(Settings)

const Container = styled.div`
  display: flex;
`

const Content = styled.div`
  flex: 1;
`

const FormContainer = styled.div`
  max-width: 350px;
`
