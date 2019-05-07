import React from 'react'
import styled from 'styled-components'
import Constants from './Constants'
import NavMenu from './NavMenu'

const LogoType = styled.h1`
  color: ${Constants.colors.dark};
  text-transform: uppercase;
`

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 2px solid ${Constants.colors.dark};
  padding: 0 20px;
  box-sizing: border-box;
`

const Header = () => (
  <Container>
    <LogoType>Tadoku</LogoType>
    <NavMenu />
  </Container>
)

export default Header
