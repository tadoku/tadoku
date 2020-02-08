import React from 'react'
import ErrorPage from 'next/error'
import { SettingsTab } from '../interfaces'
import { connect } from 'react-redux'
import { RootState } from '../../store'
import { User } from '../../session/interfaces'
import SettingsSidebar from '../components/SettingsSidebar'
import styled from 'styled-components'
import ChangePasswordForm from '../components/forms/ChangePasswordForm'
import ProfileForm from '../components/forms/ProfileForm'
import { PageTitle } from '../../ui/components'

interface Props {
  tab: SettingsTab
  user: User | undefined
  userLoaded: boolean
}

const Settings = ({ tab, user, userLoaded }: Props) => {
  if (!userLoaded) {
    return <p>Loading...</p>
  }

  if (userLoaded && !user) {
    return <ErrorPage statusCode={404} />
  }

  return (
    <>
      <PageTitle>Setting</PageTitle>
      <Container>
        <SettingsSidebar activeTab={tab} />
        <Content>
          <PageContent tab={tab} />
        </Content>
      </Container>
    </>
  )
}

const PageContent = ({ tab }: { tab: SettingsTab }) => {
  switch (tab) {
    case SettingsTab.Profile: {
      return (
        <FormContainer>
          <h2>Profile</h2>
          <ProfileForm />
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

  return null
}

const mapStateToProps = (state: RootState) => ({
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
