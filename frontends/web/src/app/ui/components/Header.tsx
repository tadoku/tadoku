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
    <InnerContainer>
      <Link href="/" passHref>
        <a href="">
          <LogoType>Tadoku</LogoType>
        </a>
      </Link>
      <ActivityIndicator isLoading={isLoading} />
      <NavigationBar />
    </InnerContainer>
  </Container>
)

const mapStateToProps = (state: RootState) => ({
  isLoading: state.app.isLoading,
})

export default connect(mapStateToProps)(Header)

const LogoType = styled.h1`
  color: ${Constants.colors.dark};
  text-transform: uppercase;

  ${media.lessThan('medium')`
    margin: 10px;
  `}
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
