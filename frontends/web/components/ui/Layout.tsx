import React from 'react'
import Header from './Header'
import styled, { createGlobalStyle } from 'styled-components'
import Constants from './Constants'

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
  }
`

const Container = styled.div`
  padding: 20px;
`

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
