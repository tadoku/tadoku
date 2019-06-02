import React, { FormEvent, useState } from 'react'
import SessionApi from '../api'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import * as SessionStore from '../redux'
import { User } from '../../user/interfaces'
import { storeUserInLocalStorage } from '../storage'
import { Form, Label, LabelText, Input, Group } from '../../ui/components/Form'
import { Button, StackContainer } from '../../ui/components'

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
      // @TODO: handle sad path
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
      <Group>
        <Label>
          <LabelText>Email</LabelText>
          <Input
            type="email"
            placeholder="tadoku@example.com"
            value={email}
            onChange={e => setEmail(e.target.value)}
          />
        </Label>
      </Group>
      <Group>
        <Label>
          <LabelText>Nickname</LabelText>
          <Input
            type="text"
            placeholder="Bob The Reader"
            value={displayName}
            onChange={e => setDisplayName(e.target.value)}
          />
        </Label>
      </Group>
      <Group>
        <Label>
          <LabelText>Password</LabelText>
          <Input
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
          />
        </Label>
      </Group>
      <Group>
        <StackContainer>
          <Button type="submit" primary>
            Create account
          </Button>
        </StackContainer>
      </Group>
    </Form>
  )
}

const mapDispatchToProps = (dispatch: Dispatch<SessionStore.Action>) => ({
  setUser: (token: string, user: User) => {
    const payload = { token, user }
    storeUserInLocalStorage(payload)

    dispatch({
      type: SessionStore.ActionTypes.SessionSignIn,
      payload,
    })
  },
})

export default connect(
  null,
  mapDispatchToProps,
)(RegisterForm)
