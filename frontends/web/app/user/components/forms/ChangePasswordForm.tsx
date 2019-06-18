import React, { FormEvent, useState } from 'react'
import {
  Form,
  Label,
  LabelText,
  Input,
  Group,
  ErrorMessage,
  GroupError,
  SuccessMessage,
} from '../../../ui/components/Form'
import { Button, ButtonContainer } from '../../../ui/components'
import { validatePassword } from '../../../session/domain'
import UserApi from '../../api'

const ChangePasswordForm = () => {
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [newPasswordConfirmation, setNewPasswordConfirmation] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(undefined as string | undefined)
  const [message, setMessage] = useState(undefined as string | undefined)

  const validate = () =>
    validatePassword(currentPassword) &&
    validatePassword(newPassword) &&
    newPassword === newPasswordConfirmation

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    setSubmitting(true)

    const response = await UserApi.changePassword({
      currentPassword,
      newPassword,
    })

    setSubmitting(false)

    if (!response) {
      setMessage(undefined)
      setError('Old password is incorrect.')
      return
    }

    setCurrentPassword('')
    setNewPassword('')
    setNewPasswordConfirmation('')
    setError(undefined)
    setMessage('Your password has been changed.')
  }

  const hasError = {
    form: !validate(),
    currentPassword:
      currentPassword != '' && !validatePassword(currentPassword),
    newPassword: newPassword != '' && !validatePassword(newPassword),
    newPasswordConfirmation:
      newPasswordConfirmation != '' && newPassword !== newPasswordConfirmation,
  }

  return (
    <Form onSubmit={submit}>
      <ErrorMessage message={error} />
      <SuccessMessage message={message} />
      <Group>
        <Label>
          <LabelText>Old password</LabelText>
          <Input
            type="password"
            value={currentPassword}
            onChange={e => setCurrentPassword(e.target.value)}
            error={hasError.currentPassword}
          />
          <GroupError
            message="Password should be at least 6 characters"
            hidden={!hasError.currentPassword}
          />
        </Label>
      </Group>
      <Group>
        <Label>
          <LabelText>New password</LabelText>
          <Input
            type="password"
            value={newPassword}
            onChange={e => setNewPassword(e.target.value)}
            error={hasError.newPassword}
          />
        </Label>
        <GroupError
          message="Password should be at least 6 characters"
          hidden={!hasError.newPassword}
        />
      </Group>
      <Group>
        <Label>
          <LabelText>New password confirmation</LabelText>
          <Input
            type="password"
            value={newPasswordConfirmation}
            onChange={e => setNewPasswordConfirmation(e.target.value)}
            error={hasError.newPasswordConfirmation}
          />
        </Label>
        <GroupError
          message="Password confirmation should be identical"
          hidden={!hasError.newPasswordConfirmation}
        />
      </Group>
      <Group>
        <ButtonContainer noMargin>
          <Button
            type="submit"
            disabled={hasError.form || submitting}
            loading={submitting}
          >
            Update password
          </Button>
        </ButtonContainer>
      </Group>
    </Form>
  )
}
export default ChangePasswordForm
