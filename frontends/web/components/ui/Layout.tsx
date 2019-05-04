import React from 'react'
import Header from './Header'
import { createGlobalStyle } from 'styled-components'
import Constants from './Constants'

const GlobalStyle = createGlobalStyle<typeof Constants>`
  body {
    background: ${props => props.colors.light};
  }
`

const Layout: React.SFC<{}> = ({ children }) => {
  return (
    <div>
      <GlobalStyle {...Constants} />
      <Header>Tadoku</Header>
      {children}
    </div>
  )
}

export default Layout
