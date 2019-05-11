import React, { FormEvent, useState } from 'react'
import styled from 'styled-components'
import Constants from '../ui/Constants'
import SessionApi from '../../domain/api/session'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import { SessionActionTypes, SessionActions } from '../../store'
import { User } from '../../domain/User'
import { storeUserInLocalStorage } from '../../domain/Session'

const Form = styled.form``
const Label = styled.label`
  display: block;
  margin-bottom: 10px;
`
const LabelText = styled.span`
  display: block;
`
const Input = styled.input`
  border: none;
  background: ${Constants.colors.secondary};
  padding: 8px 20px;
  font-size: 1.1em;
  height: 36px;
`

const Button = styled.button`
  border: none;
  background: ${Constants.colors.secondary};
  padding: 8px 20px;
  font-size: 1.1em;
  height: 36px;
`

interface Props {
  setUser: (token: string, user: User) => void
}

const SignInForm = ({ setUser }: Props) => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    // TODO: add validation
    const response = await SessionApi.signIn({ email, password })

    if (response) {
      setUser(response.token, response.user)
    }
  }

  return (
    <Form onSubmit={submit}>
      <Label>
        <LabelText>Email</LabelText>
        <Input
          type="email"
          placeholder="tadoku@example.com"
          value={email}
          onChange={e => setEmail(e.target.value)}
        />
      </Label>
      <Label>
        <LabelText>Password</LabelText>
        <Input
          type="password"
          value={password}
          onChange={e => setPassword(e.target.value)}
        />
      </Label>
      <Button type="submit">Sign in</Button>
    </Form>
  )
}

const mapDispatchToProps = (dispatch: Dispatch<SessionActions>) => ({
  setUser: (token: string, user: User) => {
    const payload = { token, user }
    storeUserInLocalStorage(payload)

    dispatch({
      type: SessionActionTypes.SessionSignIn,
      payload,
    })
  },
})

export default connect(
  null,
  mapDispatchToProps,
)(SignInForm)
