import React, { FormEvent, useState } from 'react'
import { Contest } from '../../interfaces'
import {
  Form,
  Group,
  Label,
  LabelText,
  Input,
  RadioButton,
  GroupError,
  ErrorMessage,
} from '@app/ui/components/Form'
import { Button, StackContainer } from '@app/ui/components'
import { prettyDateInUTC } from '@app/dates'
import ContestApi from '@app/contest/api'

interface Props {
  contest?: Contest
  onSuccess: () => void
  onCancel: () => void
}

const validString = (str: string | undefined): string => {
  if (!str) {
    return ''
  }
  return str
}
const validDate = (date: string | undefined): Date => {
  if (!date) {
    return new Date(0)
  }
  return new Date(date)
}

const ContestForm = ({
  contest,
  onSuccess: completed,
  onCancel: cancel,
}: Props) => {
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(undefined as string | undefined)
  const [start, setStart] = useState(() => {
    if (contest) {
      return prettyDateInUTC(contest.start)
    }

    return ''
  })
  const [end, setEnd] = useState(() => {
    if (contest) {
      return prettyDateInUTC(contest.end)
    }

    return ''
  })
  const [description, setDescription] = useState(() => {
    if (contest) {
      return contest.description || ''
    }

    return ''
  })
  const [open, setOpen] = useState(() => {
    if (contest) {
      return contest.open
    }

    return false
  })

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setSubmitting(true)

    let success: boolean

    switch (contest) {
      case undefined:
        success = await ContestApi.create({
          start: validDate(start),
          end: validDate(end),
          description: validString(description),
          open: open,
        })
        break
      default:
        success = await ContestApi.update(contest.id, {
          start: validDate(start),
          end: validDate(end),
          description: validString(description),
          open: open,
        })
        break
    }

    setSubmitting(false)

    if (!success) {
      setError('Something went wrong with saving, please try again later.')
    }

    completed()
  }

  // TODO: add proper validation
  const validate = () => start !== '' && end !== '' && description !== ''

  const hasError = {
    form: !validate(),
    start: start == '',
    end: end == '',
    description: description == '',
  }

  return (
    <Form onSubmit={submit}>
      <ErrorMessage message={error} />
      <Group>
        <Label>
          <LabelText>Description</LabelText>
          <Input
            type="text"
            value={description}
            onChange={e => setDescription(e.target.value)}
            error={hasError.description}
          />
          <GroupError
            message="Invalid description"
            hidden={!hasError.description}
          />
        </Label>
      </Group>
      <Group>
        <Label>
          <LabelText>Start</LabelText>
          <Input
            type="date"
            value={start}
            onChange={e => setStart(e.target.value)}
            error={hasError.start}
          />
          <GroupError message="Invalid start date" hidden={!hasError.start} />
        </Label>
      </Group>
      <Group>
        <Label>
          <LabelText>End</LabelText>
          <Input
            type="date"
            value={end}
            onChange={e => setEnd(e.target.value)}
            error={hasError.end}
          />
          <GroupError message="Invalid end date" hidden={!hasError.end} />
        </Label>
      </Group>
      <Group>
        <LabelText>Open</LabelText>
        <RadioButton
          type="radio"
          value="false"
          checked={open === false}
          onChange={() => setOpen(false)}
          label={'No'}
        />
        <RadioButton
          type="radio"
          value="true"
          checked={open === true}
          onChange={() => setOpen(true)}
          label={'Yes'}
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
            Save
          </Button>
          <Button type="button" onClick={cancel}>
            Cancel
          </Button>
        </StackContainer>
      </Group>
    </Form>
  )
}

export default ContestForm
