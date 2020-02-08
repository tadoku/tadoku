import React, { FormEvent, useState } from 'react'
import SessionApi from '../../api'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import { logIn } from '../../redux'
import { User } from '../../interfaces'
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
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(undefined as string | undefined)

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    setSubmitting(true)

    const success = await SessionApi.register({ email, password, displayName })

    if (!success) {
      setSubmitting(false)
      setError('Email already in use or invalid.')
      return
    }

    const response = await SessionApi.logIn({ email, password })

    setSubmitting(false)

    if (response) {
      setUser(response.token, response.user)
      complete()
    }
  }

  const validate = () =>
    validateEmail(email) &&
    validatePassword(password) &&
    validateDisplayName(displayName)

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
            message="Should be between 2-18 letters or numbers"
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
          <Button
            type="submit"
            disabled={hasError.form || submitting}
            loading={submitting}
            primary
          >
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

const mapDispatchToProps = (dispatch: Dispatch) => ({
  setUser: (token: string, user: User) => {
    const payload = { token, user }
    storeUserInLocalStorage(payload)

    dispatch(logIn(payload))
  },
})

export default connect(null, mapDispatchToProps)(RegisterForm)
