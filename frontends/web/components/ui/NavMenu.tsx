import React from 'react'
import Link from 'next/link'
import styled from 'styled-components'
import { connect } from 'react-redux'
import { State } from '../../store'
import { User } from '../../domain/User'

const StyledNav = styled.nav`
  display: flex;
  align-items: center;
`

const NavLink = styled.a`
  padding: 10px;
  display: block;
`

const LoggedInNavigation = () => (
  <>
    <Link href="/sign-out">
      <NavLink href="">Sign out</NavLink>
    </Link>
  </>
)

const LoggedOutNavigation = () => (
  <>
    <Link href="/sign-in">
      <NavLink href="">Sign in</NavLink>
    </Link>
    <Link href="/register">
      <NavLink href="">Register</NavLink>
    </Link>
  </>
)

const NavMenu = ({ user }: { user: User | undefined }) => (
  <StyledNav>
    {user ? <LoggedInNavigation /> : <LoggedOutNavigation />}
  </StyledNav>
)

const mapStateToProps = (state: State) => ({
  user: state.user,
})

export default connect(mapStateToProps)(NavMenu)
