import React from 'react'
import Link from 'next/link'
import styled from 'styled-components'

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
    <Link href="/sign-in">
      <NavLink href="">Sign in</NavLink>
    </Link>
    <Link href="/register">
      <NavLink href="">Register</NavLink>
    </Link>
  </StyledNav>
)

export default NavMenu
