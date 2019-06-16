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
import { validateDisplayName } from '../../../session/domain'
import UserApi from '../../api'

const ProfileForm = () => {
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState(undefined as string | undefined)
  const [message, setMessage] = useState(undefined as string | undefined)

  const validate = () => validateDisplayName(displayName)

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    const response = await UserApi.updateProfile({
      displayName,
    })

    if (!response) {
      setMessage(undefined)
      setError("Your profile couldn't be update. Please try again later.")
      return
    }

    setError(undefined)
    setMessage('Your profile has been updated.')
  }

  const hasError = {
    form: !validate(),
    displayName: displayName != '' && !validateDisplayName(displayName),
  }

  return (
    <Form onSubmit={submit}>
      <ErrorMessage message={error} />
      <SuccessMessage message={message} />
      <Group>
        <Label>
          <LabelText>Nickname</LabelText>
          <Input
            type="text"
            value={displayName}
            onChange={e => setDisplayName(e.target.value)}
            error={hasError.displayName}
          />
          <GroupError
            message="Display name should be at least 2 characters"
            hidden={!hasError.displayName}
          />
        </Label>
      </Group>
      <Group>
        <ButtonContainer noMargin>
          <Button type="submit" disabled={hasError.form}>
            Update profile
          </Button>
        </ButtonContainer>
      </Group>
    </Form>
  )
}
export default ProfileForm
