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
import SessionApi from '../../../session/api'
import { connect } from 'react-redux'
import { logIn } from '../../../session/redux'
import { User } from '../../../session/interfaces'
import { storeUserInLocalStorage } from '../../../session/storage'
import { RootState, Dispatch } from '../../../store'

interface Props {
  setUser: (token: string, user: User) => void
  user: User | undefined
  userLoaded: boolean
}

const ProfileForm = ({ user, setUser, userLoaded }: Props) => {
  const [displayName, setDisplayName] = useState(() =>
    user ? user.displayName : '',
  )
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(undefined as string | undefined)
  const [message, setMessage] = useState(undefined as string | undefined)

  if (userLoaded && !user) {
    return null
  }

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    setSubmitting(true)

    const response = await UserApi.updateProfile({
      displayName,
    })

    setSubmitting(false)

    if (!response) {
      setMessage(undefined)
      setError("Your profile couldn't be update. Please try again later.")
      return
    }

    setError(undefined)
    setMessage('Your profile has been updated.')

    const sessionResponse = await SessionApi.refresh()
    if (sessionResponse) {
      setUser(sessionResponse.token, sessionResponse.user)
    }
  }

  const validate = () => validateDisplayName(displayName)

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
            message="Should be between 2-18 letters or numbers"
            hidden={!hasError.displayName}
          />
        </Label>
      </Group>
      <Group>
        <ButtonContainer noMargin>
          <Button
            type="submit"
            disabled={hasError.form || submitting}
            loading={submitting}
          >
            Update profile
          </Button>
        </ButtonContainer>
      </Group>
    </Form>
  )
}

const mapStateToProps = (state: RootState) => ({
  user: state.session.user,
  userLoaded: state.session.loaded,
})

const mapDispatchToProps = (dispatch: Dispatch) => ({
  setUser: (token: string, user: User) => {
    const payload = { token, user }
    storeUserInLocalStorage(payload)
    dispatch(logIn(payload))
  },
})

export default connect(mapStateToProps, mapDispatchToProps)(ProfileForm)
