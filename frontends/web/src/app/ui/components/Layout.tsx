import React from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import media from 'styled-media-query'
import { useSelector } from 'react-redux'

import Header from './Header'
import Constants from '../Constants'
import Footer from './Footer'
import ActivityIndicator from './ActivityIndicator'
import { RootState } from '../../store'
import { ContestMapper } from '../../contest/transform'

interface Props {
  overridesLayout: boolean
}

const Layout: React.SFC<Props> = ({ children, overridesLayout }) => {
  const isLoading = useSelector((state: RootState) => state.app.isLoading)
  const contests = useSelector((state: RootState) =>
    state.contest.recentContests.map(ContestMapper.fromRaw),
  )

  if (overridesLayout) {
    return (
      <div>
        <GlobalStyle {...Constants} />
        {children}
      </div>
    )
  }

  return (
    <div>
      <ActivityIndicator isLoading={isLoading} />
      <GlobalStyle {...Constants} />
      <StickyFooterContainer>
        <Header />
        <Container>{children}</Container>
      </StickyFooterContainer>
      <Footer contests={contests} />
    </div>
  )
}

export default Layout

export const GlobalStyle = createGlobalStyle<typeof Constants>`
  html,
  body {
    position: relative;
  }

  html {
    height: 100%;
    overflow-x: hidden;
    margin-right: calc(-1 * (100vw - 100%));
  }

  body {
    background: ${props => props.colors.lighter};
    font-family: ${props => props.fonts.sansSerif};
    margin: 0;
    padding: 0;
  }

  ::selection {
    background: ${props => props.colors.primary};
    color: ${props => props.colors.light};
  }

  a {
    color: ${props => props.colors.dark};
    text-decoration: none;
    transition: color 0.2s ease;

    &:hover, &:active, &:focus {
      color: ${props => props.colors.primary}
    }
  }

  h1, h2, h3, h4, h5, h6, h7 {
    font-family: ${props => props.fonts.serif};
  }

  a[href],
  input[type='submit']:not([disabled]),
  input[type='image']:not([disabled]),
  label[for]:not([disabled]),
  select:not([disabled]),
  button:not([disabled]) {
    cursor: pointer;
  }
`

const Container = styled.div`
  box-sizing: border-box;
  max-width: ${Constants.maxWidth};
  margin: 50px auto 60px;
  padding: 0 30px;
  box-sizing: border-box;

  ${media.lessThan('medium')`
    padding: 0 20px;
  `}
`

const StickyFooterContainer = styled.div`
  min-height: 100vh;
  overflow: hidden;
  position: relative;
  box-sizing: border-box;

  ${media.greaterThan('medium')`
    padding-bottom: 250px;
  `}
`
