import React from 'react'
import styled, { createGlobalStyle } from 'styled-components'

import Header from './../components/Header'

const LandingPage = () => {
  return (
    <Container>
      <GlobalStyle />
      <Header />
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
