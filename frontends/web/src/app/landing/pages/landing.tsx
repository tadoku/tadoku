import React from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import { useDispatch } from 'react-redux'

import Header from './../components/Header'
import * as RankingStore from '../../ranking/redux'

const LandingPage = () => {
  const dispatch = useDispatch()
  const refreshSession = () => {
    dispatch({
      type: RankingStore.ActionTypes.RankingRunEffects,
    })
  }

  return (
    <Container>
      <GlobalStyle />
      <Header refreshSession={refreshSession} />
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
