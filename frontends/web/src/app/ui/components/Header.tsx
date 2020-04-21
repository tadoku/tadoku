import React, { useState } from 'react'
import { RootState } from '../../store'
import { connect } from 'react-redux'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import Link from 'next/link'
import styled from 'styled-components'
import media from 'styled-media-query'
import Constants from '../Constants'
import NavigationBar from './navigation/NavigationBar'
import ActivityIndicator from './ActivityIndicator'

interface Props {
  isLoading: boolean
}

const Header = ({ isLoading }: Props) => {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <Container>
      <Link href="/" passHref>
        <LogoLink href="">
          <Logo />
        </LogoLink>
      </Link>
      <ActivityIndicator isLoading={isLoading} />
      <NavigationBar isOpen={isOpen} />
      <Hamburger onClick={() => setIsOpen(!isOpen)}>
        <FontAwesomeIcon
          icon={isOpen ? 'times' : 'bars'}
          rotation={isOpen ? 90 : undefined}
          size="2x"
        />
      </Hamburger>
    </Container>
  )
}

const mapStateToProps = (state: RootState) => ({
  isLoading: state.app.isLoading,
})

export default connect(mapStateToProps)(Header)

const Logo = styled.img.attrs(() => ({
  src: '/img/logo.svg',
  alt: 'Tadoku',
}))`
  height: 29px;
  width: 158px;
`

const Hamburger = styled.div`
  position: absolute;
  top: 30px;
  right: 30px;
  height: 29px;
  width: 29px;
  display: none;

  svg {
    transition: 0.2s all ease-out;
    max-height: 29px;
    max-width: 33px;
  }

  ${media.lessThan('medium')`
    display: block;
  `}
`

const Container = styled.div`
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  background: ${Constants.colors.lightTinted};
  height: 120px;
  max-width: ${Constants.maxWidth};
  box-sizing: border-box;
  padding: 0 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 auto;

  ${media.lessThan('medium')`
    padding: 30px;
    height: inherit;
    align-items: start;
    flex-direction: column;
  `}
`

const LogoLink = styled.a`
  display: block;
  height: 29px;
`
