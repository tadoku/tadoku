import React, { FormEvent, useState } from 'react'
import SessionApi from '../../api'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import * as SessionStore from '../../redux'
import { User } from '../../../user/interfaces'
import { storeUserInLocalStorage } from '../../storage'
import {
  Form,
  Label,
  LabelText,
  Input,
  Group,
  ErrorMessage,
  GroupError,
} from '../../../ui/components/Form'
import { Button, StackContainer } from '../../../ui/components'
import {
  validateEmail,
  validatePassword,
  validateDisplayName,
} from '../../domain'

interface Props {
  setUser: (token: string, user: User) => void
  onSuccess: () => void
  onCancel: () => void
}

const RegisterForm = ({
  setUser,
  onSuccess: complete,
  onCancel: cancel,
}: Props) => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState(undefined as string | undefined)

  const validate = () =>
    validateEmail(email) &&
    validatePassword(password) &&
    validateDisplayName(displayName)

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    // TODO: add validation
    const success = await SessionApi.register({ email, password, displayName })

    if (!success) {
      // @TODO: handle sad path
      setError('Email already in use or invalid.')
      return
    }

    const response = await SessionApi.logIn({ email, password })

    if (response) {
      setUser(response.token, response.user)
      setError(undefined)
      complete()
    }
  }

  const hasError = {
    form: !validate(),
    email: email !== '' && !validateEmail(email),
    displayName: displayName !== '' && !validateDisplayName(displayName),
    password: password != '' && !validatePassword(password),
  }

  return (
    <Form onSubmit={submit}>
      <Group>
        <ErrorMessage message={error} />
        <Label>
          <LabelText>Email</LabelText>
          <Input
            type="email"
            placeholder="tadoku@example.com"
            value={email}
            onChange={e => setEmail(e.target.value)}
            error={hasError.email}
          />
          <GroupError message="Invalid email" hidden={!hasError.email} />
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
            error={hasError.displayName}
          />
          <GroupError
            message="Display name should be at least 6 characters"
            hidden={!hasError.displayName}
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
            error={hasError.password}
          />
          <GroupError
            message="Password should be at least 6 characters"
            hidden={!hasError.password}
          />
        </Label>
      </Group>
      <Group>
        <StackContainer>
          <Button type="submit" primary disabled={hasError.form}>
            Create account
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
)(RegisterForm)
