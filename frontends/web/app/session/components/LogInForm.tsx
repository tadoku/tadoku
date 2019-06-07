import React, { FormEvent, useState } from 'react'
import SessionApi from '../api'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import * as SessionStore from '../redux'
import { User } from '../../user/interfaces'
import { storeUserInLocalStorage } from '../storage'
import {
  Form,
  Label,
  LabelText,
  Input,
  Group,
  ErrorMessage,
} from '../../ui/components/Form'
import { Button, StackContainer } from '../../ui/components'
import { validatePassword, validateEmail } from '../domain'

interface Props {
  setUser: (token: string, user: User) => void
  onSuccess: () => void
  onCancel: () => void
}

const LogInForm = ({
  setUser,
  onSuccess: complete,
  onCancel: cancel,
}: Props) => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState(undefined as string | undefined)

  const validate = () => validateEmail(email) && validatePassword(password)

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    const response = await SessionApi.logIn({ email, password })

    if (!response) {
      setError('Invalid email/password combination.')
      return
    }

    setUser(response.token, response.user)
    setError(undefined)
    complete()
  }

  const valid = validate()

  return (
    <Form onSubmit={submit}>
      <ErrorMessage message={error} />
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
          <Button type="submit" primary disabled={!valid}>
            Sign in
          </Button>
          <Button type="button" onClick={cancel}>
            Cancel
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
      type: SessionStore.ActionTypes.SessionLogIn,
      payload,
    })
  },
})

export default connect(
  null,
  mapDispatchToProps,
)(LogInForm)
