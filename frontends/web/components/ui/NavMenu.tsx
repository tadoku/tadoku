import React from 'react'
import styled from 'styled-components'
import Constants from './Constants'

const StyledNav = styled.nav`
  display: flex;
  align-items: center;
`

const NavLink = styled.a`
  padding: 10px;
  display: block;
`

const NavMenu = () => (
  <StyledNav>
    <NavLink href="">Sign in</NavLink>
    <NavLink href="">Register</NavLink>
  </StyledNav>
)

export default NavMenu
