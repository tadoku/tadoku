import React, { useState } from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import { useDispatch } from 'react-redux'

import Header from './../components/Header'
import LogInModal from './../../session/components/modals/LogInModal'
import * as RankingStore from '../../ranking/redux'

const LandingPage = () => {
  const dispatch = useDispatch()
  const refreshSession = () => {
    dispatch({
      type: RankingStore.ActionTypes.RankingRunEffects,
    })
  }

  const [isLoginModalOpen, setIsLoginModalOpen] = useState(false)

  return (
    <Container>
      <GlobalStyle />
      <LogInModal
        isOpen={isLoginModalOpen}
        onSuccess={refreshSession}
        onCancel={() => setIsLoginModalOpen(false)}
      />
      <Header
        refreshSession={refreshSession}
        openLoginModal={() => setIsLoginModalOpen(true)}
      />
    </Container>
  )
}

export default LandingPage

export const GlobalStyle = createGlobalStyle`
  background: hsl(0, 0, 98);
`

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`
