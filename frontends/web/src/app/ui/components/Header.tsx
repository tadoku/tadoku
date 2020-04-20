import React from 'react'
import { RootState } from '../../store'
import { connect } from 'react-redux'
import Link from 'next/link'
import styled from 'styled-components'
import media from 'styled-media-query'
import Constants from '../Constants'
import NavigationBar from './navigation/NavigationBar'
import ActivityIndicator from './ActivityIndicator'

interface Props {
  isLoading: boolean
}

const Header = ({ isLoading }: Props) => (
  <Container>
    <Link href="/" passHref>
      <a href="">
        <Logo />
      </a>
    </Link>
    <ActivityIndicator isLoading={isLoading} />
    <NavigationBar />
  </Container>
)

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
    flex-direction: column;
    box-shadow: none;
  `}
`
