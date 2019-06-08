import React from 'react'
import Link from 'next/link'
import styled from 'styled-components'
import Constants from '../Constants'
import NavigationBar from './navigation/NavigationBar'

const Header = () => (
  <Container>
    <Link href="/">
      <a href="">
        <LogoType>Tadoku</LogoType>
      </a>
    </Link>
    <NavigationBar />
  </Container>
)

export default Header

const LogoType = styled.h1`
  color: ${Constants.colors.dark};
  text-transform: uppercase;
`

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 4px 3px 7px 1px rgba(0, 0, 0, 0.08);
  padding: 0 20px;
  box-sizing: border-box;
`
