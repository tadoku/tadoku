import React, { useState } from 'react'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import Link from 'next/link'
import styled from 'styled-components'
import media from 'styled-media-query'
import Constants from '../Constants'
import NavigationBar from './navigation/NavigationBar'
import { Logo } from './index'

interface Props {
  homeURL: string
}

const Header = ({ homeURL }: Props) => {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <Container>
      <InnerContainer>
        <Link href={homeURL} passHref>
          <LogoLink href="">
            <Logo />
          </LogoLink>
        </Link>
        <NavigationBar isOpen={isOpen} />
        <Hamburger onClick={() => setIsOpen(!isOpen)}>
          <FontAwesomeIcon
            icon={isOpen ? 'times' : 'bars'}
            rotation={isOpen ? 90 : undefined}
            size="2x"
          />
        </Hamburger>
      </InnerContainer>
    </Container>
  )
}

export default Header

const Hamburger = styled.div`
  position: absolute;
  top: 20px;
  right: 20px;
  height: 50px;
  width: 50px;
  display: none;
  padding: 10px;
  box-sizing: border-box;

  svg {
    transition: 0.2s all ease-out;
    max-height: 29px;
    max-width: 33px;
  }

  ${media.lessThan('medium')`
    display: block;
  `}

  &:hover {
    cursor: pointer;
  }
`

const InnerContainer = styled.div`
  max-width: ${Constants.maxWidth};
  height: 100%;
  box-sizing: border-box;
  padding: 0 30px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 auto;

  ${media.lessThan('medium')`
    padding: 30px 20px;
    align-items: start;
    flex-direction: column;
  `}
`

const Container = styled.div`
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  background: ${Constants.colors.light};
  height: 120px;
  width: 100%;

  ${media.lessThan('medium')`
    height: inherit;
  `}
`

const LogoLink = styled.a`
  display: block;
  height: 29px;
`
