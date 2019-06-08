import React from 'react'
import Header from './Header'
import styled, { createGlobalStyle } from 'styled-components'
import Constants from '../Constants'

const Layout: React.SFC<{}> = ({ children }) => {
  return (
    <div>
      <GlobalStyle {...Constants} />
      <Header />
      <Container>{children}</Container>
    </div>
  )
}

export default Layout

const GlobalStyle = createGlobalStyle<typeof Constants>`
  body {
    background: ${props => props.colors.light};
    font-family: 'Open Sans', sans-serif;
    margin: 0;
    padding: 0;
  }

  a {
    color: ${props => props.colors.dark}
    text-decoration: none;

    &:hover, &:active, &:focus {
      color: ${props => props.colors.primary}
    }
  }
`

const Container = styled.div`
  padding: 20px;
`
