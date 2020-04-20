import React from 'react'
import Link from 'next/link'
import styled from 'styled-components'
import media from 'styled-media-query'
import Constants from '../Constants'

const Footer = () => (
  <Container>
    <InnerContainer>
      <Link href="/" passHref>
        <a href="">
          <LogoType>Tadoku</LogoType>
        </a>
      </Link>
      <Credits>
        Source code available on <a href="https://github.com/tadoku">GitHub</a>
        <br />
        Built by <a href="https://antonve.be">antonve</a>
      </Credits>
    </InnerContainer>
  </Container>
)

export default Footer

const LogoType = styled.h4`
  color: ${Constants.colors.dark};
  text-transform: uppercase;

  ${media.lessThan('medium')`
    margin: 10px;
  `}
`

const Container = styled.div`
  box-sizing: border-box;
  height: 100px;
  display: none;

  ${media.greaterThan('medium')`
      display: block;
      position: absolute;
      left: 0;
      right: 0;
      bottom: 0;
  `}
`

const InnerContainer = styled.div`
  border-top: 1px solid ${Constants.colors.lightGray};
  max-width: ${Constants.maxWidth};
  display: flex;
  align-items: top;
  justify-content: space-between;
  margin: 0 auto;
  padding: 0 60px;
  box-sizing: border-box;

  ${media.lessThan('large')`
    padding: 0 20px;
  `}
`

const Credits = styled.p`
  text-align: right;
  line-height: 2em;

  a {
    display: inline-block;
    border-bottom: 2px solid ${Constants.colors.primary};
    height: 1.6em;
  }
`
