import React, { FormEvent, useState, useEffect } from 'react'
import { AllMediums, languageNameByCode } from '../database'
import { connect } from 'react-redux'
import { State } from '../../store'
import { RankingRegistration, ContestLog } from '../interfaces'
import RankingApi from '../api'
import {
  Form,
  Group,
  Label,
  LabelText,
  Input,
  Select,
  LabelForRadio,
} from '../../ui/components/Form'
import { Button, StackContainer } from '../../ui/components'

interface Props {
  log?: ContestLog
  registration?: RankingRegistration | undefined
  onSuccess: () => void
  onCancel: () => void
}

const EditLogForm = ({
  log,
  registration,
  onSuccess: completed,
  onCancel: cancel,
}: Props) => {
  const [amount, setAmount] = useState(() => {
    if (log) {
      return log.amount.toString()
    }

    return ''
  })
  const [mediumId, setMediumId] = useState(() => {
    if (log) {
      return log.mediumId.toString()
    }

    return ''
  })
  const [languageCode, setLanguageCode] = useState(() => {
    if (log) {
      return log.languageCode
    }

    return ''
  })
  const [changed, setChanged] = useState(false)
  const [isFirstRun, setIsFirstRun] = useState(true)

  useEffect(() => {
    if (!isFirstRun) {
      setChanged(true)
    }
    setIsFirstRun(false)
  }, [amount, mediumId, languageCode])

  if (!registration) {
    return null
  }

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    let success: boolean

    switch (log) {
      case undefined:
        success = await RankingApi.createLog({
          contestId: registration.contestId,
          mediumId: Number(mediumId),
          amount: Number(amount),
          languageCode,
        })
        break
      default:
        success = await RankingApi.updateLog(log.id, {
          contestId: log.contestId,
          mediumId: Number(mediumId),
          amount: Number(amount),
          languageCode,
        })
        break
    }

    if (success) {
      completed()
    }
  }

  return (
    <Form onSubmit={submit}>
      <Group>
        <Label>
          <LabelText>Pages read</LabelText>
          <Input
            type="number"
            placeholder="e.g. 7"
            value={amount}
            onChange={e => setAmount(e.target.value)}
            min={0}
            max={3000}
            step={1}
          />
        </Label>
      </Group>
      <Group>
        <Label>
          <LabelText>Medium</LabelText>
          <Select value={mediumId} onChange={e => setMediumId(e.target.value)}>
            {AllMediums.map(m => (
              <option value={m.id} key={m.id}>
                {m.description}
              </option>
            ))}
          </Select>
        </Label>
      </Group>
      <Group>
        <LabelText>Language</LabelText>
        {registration.languages.map(code => (
          <LabelForRadio key={code}>
            <input
              type="radio"
              value={code}
              checked={code === languageCode}
              onChange={e => setLanguageCode(e.target.value)}
            />
            <span>{languageNameByCode(code)}</span>
          </LabelForRadio>
        ))}
      </Group>
      <Group>
        <StackContainer>
          <Button type="submit" disabled={!changed} primary>
            Save changes
          </Button>
          <Button onClick={cancel}>Cancel</Button>
        </StackContainer>
      </Group>
    </Form>
  )
}

const mapStateToProps = (state: State, oldProps: Props) => ({
  ...oldProps,
  registration: state.ranking.registration,
})

export default connect(mapStateToProps)(EditLogForm)
