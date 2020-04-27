import React, { useState } from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import { useDispatch, useSelector } from 'react-redux'
import media from 'styled-media-query'

import Header from './../components/Header'
import ContestSchedule from '../components/ContestSchedule'
import LogInModal from './../../session/components/modals/LogInModal'
import * as RankingStore from '../../ranking/redux'
import { FooterLanding } from '../../ui/components/Footer'
import Constants from '../../ui/Constants'
import { RootState } from '../../store'
import { contestMapper } from '../../contest/transform'

const LandingPage = () => {
  const dispatch = useDispatch()
  const refreshSession = () => {
    dispatch(RankingStore.runEffects())
  }

  const [isLoginModalOpen, setIsLoginModalOpen] = useState(false)
  const contests = useSelector((state: RootState) =>
    state.contest.recentContests.map(contestMapper.fromRaw),
  )

  return (
    <Container>
      <GlobalStyle />
      <LogInModal
        isOpen={isLoginModalOpen}
        onSuccess={refreshSession}
        onCancel={() => setIsLoginModalOpen(false)}
      />
      <StickyFooterContainer>
        <Header
          refreshSession={refreshSession}
          openLoginModal={() => setIsLoginModalOpen(true)}
        />
        <Content>
          <Card>
            <Title>Why should I participate?</Title>
            <Paragraph>
              Extensive reading of native materials is a great way to improve
              your understanding of the language you&apos;re learning. There are
              many benefits to doing so: it builds vocabulary, reinforces
              grammar patterns, and you learn about the culture of where your
              language is spoken. As you participate in more rounds you will
              notice that you can read more and more as you improve.
            </Paragraph>
            <Paragraph>
              That said, it&apos;s not for everyone. Not everyone enjoys the
              process of immersing themselves. Tadoku isn&apos;t a magical pill
              that will make you fluent. It only covers extensive reading, and
              not extensive listening. While Tadoku is here to promote reading,
              a balanced approach to learning is still recommended.
            </Paragraph>
          </Card>
          <ContestSchedule />
        </Content>
      </StickyFooterContainer>
      <FooterLanding contests={contests} />
    </Container>
  )
}

export default LandingPage

export const GlobalStyle = createGlobalStyle`
  background: hsl(0, 0, 98);
`

const Container = styled.div``

const Content = styled.div`
  margin: 90px auto;
  width: 1200px;

  ${media.lessThan('large')`
    width: 100%;
  `}
`

const Title = styled.h2`
  font-family: ${Constants.fonts.serif};
  margin: 0 0 30px 0;
  font-size: 24px;
  line-height: 31px;
  font-weight: 700;
  letter-spacing: 1.05;
`

const Paragraph = styled.p`
  font-family: ${Constants.fonts.sansSerif};
  font-size: 18px;
  line-height: 29px;
`

const Card = styled.div`
  max-width: 695px;
  box-sizing: border-box;
  padding: 0 60px;

  ${media.lessThan('large')`
    max-width: 100%;
  `}
`

const StickyFooterContainer = styled.div`
  min-height: 100vh;
  overflow: hidden;
  position: relative;
  box-sizing: border-box;

  ${media.greaterThan('medium')`
    // height of the footer
    padding-bottom: 250px;
  `}
`
