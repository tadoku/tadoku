import React from 'react'
import Header from './Header'
import styled, { createGlobalStyle } from 'styled-components'
import Constants from '../Constants'
import Footer from './Footer'
import media from 'styled-media-query'

const Layout: React.SFC<{}> = ({ children }) => {
  return (
    <div>
      <GlobalStyle {...Constants} />
      <StickyFooterContainer>
        <Header />
        <Container>{children}</Container>
      </StickyFooterContainer>
      <Footer />
    </div>
  )
}

export default Layout

const GlobalStyle = createGlobalStyle<typeof Constants>`
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
    background: ${props => props.colors.light};
    font-family: 'Open Sans', sans-serif;
    margin: 0;
    padding: 0;
  }

  a {
    color: ${props => props.colors.dark}
    text-decoration: none;
    transition: color 0.2s ease;

    &:hover, &:active, &:focus {
      color: ${props => props.colors.primary}
    }
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
  padding: 20px;
  max-width: ${Constants.maxWidth};
  margin: 0 auto;
`

const StickyFooterContainer = styled.div`
  min-height: 100vh;
  overflow: hidden;
  position: relative;
  // height of the footer
  box-sizing: border-box;

  ${media.greaterThan('medium')`
    padding-bottom: 150px;
  `}
`
