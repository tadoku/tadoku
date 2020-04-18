import React, { useState } from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import { useDispatch } from 'react-redux'
import media from 'styled-media-query'

import Header from './../components/Header'
import LogInModal from './../../session/components/modals/LogInModal'
import * as RankingStore from '../../ranking/redux'
import Footer from '../../ui/components/Footer'

const LandingPage = () => {
  const dispatch = useDispatch()
  const refreshSession = () => {
    dispatch(RankingStore.runEffects())
  }

  const [isLoginModalOpen, setIsLoginModalOpen] = useState(false)

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
              notice that you can read more and more as your command of your
              target language improves.
            </Paragraph>
            <Paragraph>
              That said, it&apos;s not for everyone. Not everyone enjoys the
              process of immersing themselves. Tadoku isn&apos;t a magical pill
              that will make you fluent. It only covers extensive reading, and
              not extensive listening. While Tadoku is here to promote reading,
              a balanced approach to learning is still recommended.
            </Paragraph>
          </Card>
        </Content>
      </StickyFooterContainer>
      <Footer />
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
  width: 1240px;

  ${media.lessThan('large')`
    width: 100%;
  `}
`

const Title = styled.h2`
  font-family: 'Merriweather', serif;
  margin: 0 0 30px 0;
  font-size: 24px;
  line-height: 31px;
  font-weight: 700;
  letter-spacing: 1.05;
`

const Paragraph = styled.p`
  font-family: 'Open sans', serif;
  font-size: 18px;
  line-height: 29px;
`

const Card = styled.div`
  max-width: 715px;
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
  // height of the footer
  box-sizing: border-box;

  ${media.greaterThan('medium')`
    padding-bottom: 150px;
  `}
`
