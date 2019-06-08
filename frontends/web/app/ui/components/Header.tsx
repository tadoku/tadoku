import React from 'react'
import Link from 'next/link'
import styled from 'styled-components'
import media from 'styled-media-query'
import Constants from '../Constants'
import NavigationBar from './navigation/NavigationBar'

const Header = () => (
  <Container>
    <InnerContainer>
      <Link href="/">
        <a href="">
          <LogoType>Tadoku</LogoType>
        </a>
      </Link>
      <NavigationBar />
    </InnerContainer>
  </Container>
)

export default Header

const LogoType = styled.h1`
  color: ${Constants.colors.dark};
  text-transform: uppercase;
`

const Container = styled.div`
  box-shadow: 4px 3px 7px 1px rgba(0, 0, 0, 0.08);
  padding: 0 20px;
  box-sizing: border-box;

  ${media.lessThan('medium')`
    box-shadow: none;
  `}
`

const InnerContainer = styled.div`
  max-width: ${Constants.maxWidth};
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 auto;

  ${media.lessThan('medium')`
    flex-direction: column;
  `}
`
