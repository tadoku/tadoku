import React from 'react'
import styled from 'styled-components'

const Header = styled.h1`
  color: red;
`

const Layout: React.SFC<{}> = ({ children }) => {
  return (
    <div>
      <Header>Tadoku</Header>
      {children}
    </div>
  )
}

export default Layout
