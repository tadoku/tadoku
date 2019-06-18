import React, { FormEvent, useState } from 'react'
import SessionApi from '../../api'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import * as SessionStore from '../../redux'
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
import { validatePassword, validateEmail } from '../../domain'

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
  const [submitting, setSubmitting] = useState(false)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState(undefined as string | undefined)

  const validate = () => validateEmail(email) && validatePassword(password)

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    setSubmitting(true)

    const response = await SessionApi.logIn({ email, password })

    setSubmitting(false)

    if (!response) {
      setError('Invalid email/password combination.')
      return
    }

    setUser(response.token, response.user)
    complete()
  }

  const hasError = {
    form: !validate(),
    email: email !== '' && !validateEmail(email),
    password: password != '' && !validatePassword(password),
  }

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
            error={hasError.email}
          />
          <GroupError message="Invalid email" hidden={!hasError.email} />
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
        </Label>
        <GroupError
          message="Password should be at least 6 characters"
          hidden={!hasError.password}
        />
      </Group>
      <Group>
        <StackContainer>
          <Button
            type="submit"
            disabled={hasError.form || submitting}
            loading={submitting}
            primary
          >
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
