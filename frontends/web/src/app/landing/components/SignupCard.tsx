import React from 'react'
import styled from 'styled-components'
import RegisterForm from '@app/session/components/forms/RegisterForm'
import media from 'styled-media-query'
import Constants from '@app/ui/Constants'

interface Props {
  refreshSession: () => void
  openLoginModal: () => void
}

const SignupCard = ({ refreshSession, openLoginModal }: Props) => (
  <Container>
    <Card>
      <Large>
        <SignupTitle>
          Create a <br />
          new account
        </SignupTitle>
        <RegisterForm onSuccess={refreshSession} />
        <LoginPrompt>
          Already have an account?{' '}
          <a onClick={openLoginModal} href="#">
            Log in
          </a>
        </LoginPrompt>
      </Large>
    </Card>
  </Container>
)

export default SignupCard

const Container = styled.div`
  width: 400px;
  position: absolute;
  top: 120px;
  right: calc((100% - ${Constants.maxWidth}) / 2 + 105px);
  bottom: 340px;

  ${media.lessThan('large')`
    position: inherit;
    top: inherit;
    right: inherit;
    width: inherit;
    min-width: 400px;
    flex: 1;
  `}

  ${media.lessThan('medium')`
    max-width: 500px;
    min-width: inherit;
    width: 100%;
    margin: 0 auto;
  `}
`

const Card = styled.div`
  background: #fff;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  width: 100%;
  box-sizing: border-box;
  padding: 60px;

  ${media.lessThan('large')`
    box-shadow: none;
  `}

  @supports (position: -webkit-sticky) or (position: sticky) {
    position: sticky;
    top: 60px;
    right: 0;
    z-index: 100;

    @media screen and (max-height: 700px) {
      & {
        top: 0;
        position: absolute;
      }
    }
  }
`

const Large = styled.div`
  ${media.lessThan('large')`
    /* display: none; */
  `}
`

const SignupTitle = styled.h2`
  font-family: ${Constants.fonts.serif};
  margin: 0 0 30px 0;
  font-size: 26px;
  line-height: 32px;
  font-weight: 700;
`

const LoginPrompt = styled.p`
  font-family: ${Constants.fonts.sansSerif};
  font-size: 16px;
  text-align: center;

  a {
    color: #725dff;
    font-weight: 600;
    text-decoration: underline;
  }
`
