import React from 'react'
import styled from 'styled-components'
import RegisterForm from '../../session/components/forms/RegisterForm'
import media from 'styled-media-query'

interface Props {
  refreshSession: () => void
  openLoginModal: () => void
}

const SignupCard = ({ refreshSession, openLoginModal }: Props) => (
  <Card>
    <Large>
      <SignupTitle>
        Create a <br />
        new account
      </SignupTitle>
      <RegisterForm onSuccess={refreshSession} />
      <LoginPrompt>
        Already have an account? <a onClick={openLoginModal}>Log in</a>
      </LoginPrompt>
    </Large>
  </Card>
)

export default SignupCard

const Card = styled.div`
  background: #fff;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  width: 400px;
  box-sizing: border-box;
  padding: 60px;
  position: absolute;
  top: 120px;
  right: 105px;

  ${media.lessThan('large')`
    position: inherit;
    top: inherit;
    right: inherit;
    width: inherit;
    min-width: 400px;
    flex: 1;
    box-shadow: none;
  `}

  ${media.lessThan('medium')`
    max-width: 500px;
    min-width: inherit;
    width: 100%;
    margin: 0 auto;
  `}
`

const Large = styled.div`
  ${media.lessThan('large')`
    /* display: none; */
  `}
`

const SignupTitle = styled.h2`
  font-family: 'Merriweather', serif;
  margin: 0 0 30px 0;
  font-size: 26px;
  line-height: 32px;
  font-weight: 700;
`

const LoginPrompt = styled.p`
  font-family: 'Open sans', serif;
  font-size: 16px;
  text-align: center;

  a {
    color: #725dff;
    font-weight: 600;
    text-decoration: underline;
  }
`
