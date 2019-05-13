import React, { FormEvent, useState } from 'react'
import styled from 'styled-components'
import Constants from '../../ui/Constants'
import SessionApi from '../api'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import { SessionActionTypes, SessionActions } from '../../../store'
import { User } from '../../user/interfaces'
import { storeUserInLocalStorage } from '../storage'

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

const RegisterForm = ({ setUser }: Props) => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [displayName, setDisplayName] = useState('')

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    // TODO: add validation
    const success = await SessionApi.register({ email, password, displayName })

    if (!success) {
      // handle sad path
      console.log("shit happened, couldn't register")
      return
    }

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
        <LabelText>Nickname</LabelText>
        <Input
          type="text"
          placeholder="Bob The Reader"
          value={displayName}
          onChange={e => setDisplayName(e.target.value)}
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
      <Button type="submit">Create account</Button>
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
)(RegisterForm)
