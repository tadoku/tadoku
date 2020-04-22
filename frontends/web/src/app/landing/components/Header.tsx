import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'

import SignupCard from './SignupCard'
import { Logo } from '../../ui/components'

interface Props {
  refreshSession: () => void
  openLoginModal: () => void
}

const Header = ({ refreshSession, openLoginModal }: Props) => (
  <Background>
    <Grid>
      <IntroCard>
        <Logo />
        <Title>Get good at your second language</Title>
        <Tagline>
          Tadoku is a friendly foreign-language reading contest aimed at
          building a habit of reading in your non-native languages.
        </Tagline>
      </IntroCard>
      <SignupCard
        refreshSession={refreshSession}
        openLoginModal={openLoginModal}
      />
    </Grid>
  </Background>
)

export default Header

const Background = styled.div`
  width: 100%;
  max-width: 1850px;
  height: 460px;
  margin: 0 auto;
  background-image: url('/img/header.jpg');
  background-size: cover;

  ${media.lessThan('large')`
    height: inherit;
  `}

  ${media.lessThan('medium')`
    background: none;
  `}
`

const Grid = styled.div`
  max-width: 1200px;
  margin: 0 auto;
  position: relative;

  ${media.lessThan('large')`
    display: flex;
    align-items:flex-start;
  `}

  ${media.lessThan('medium')`
    flex-direction: column;
  `}
`

const IntroCard = styled.div`
  background: #f2f8ff;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  max-width: 505px;
  box-sizing: border-box;
  padding: 60px;

  ${media.lessThan('large')`
    box-shadow: none;
    max-height: inherit;
  `}

  ${media.lessThan('medium')`
    max-width: 100%;
  `}
`

const Title = styled.h1`
  font-family: 'Merriweather', serif;
  margin: 60px 20px 30px 0;
  font-size: 30px;
  line-height: 37px;
  font-weight: 700;
`

const Tagline = styled.p`
  font-size: 18px;
  line-height: 29px;
  font-family: 'Open sans', serif;
  padding: 0;
  margin: 0;
`
