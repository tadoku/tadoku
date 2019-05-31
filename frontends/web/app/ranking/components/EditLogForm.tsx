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
  Button,
} from '../../ui/components/Form'

interface Props {
  log: ContestLog
  registration: RankingRegistration | undefined
  onSuccess: () => void
}

const EditLogForm = ({ log, registration, onSuccess: completed }: Props) => {
  const [amount, setAmount] = useState(log.amount.toString())
  const [mediumId, setMediumId] = useState(log.mediumId.toString())
  const [languageCode, setLanguageCode] = useState(log.languageCode)
  const [changed, setChanged] = useState(false)
  const [isFirstRun, setIsFirstRun] = useState(true)

  useEffect(() => {
    if (!isFirstRun) {
      setChanged(true)
    }
    setIsFirstRun(false)
  }, [amount, mediumId, languageCode])

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    const success = await RankingApi.updateLog(log.id, {
      contestId: log.contestId,
      mediumId: Number(mediumId),
      amount: Number(amount),
      languageCode,
    })

    if (success) {
      completed()
    }
  }

  if (!registration) {
    return null
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
        <Button type="submit" disabled={!changed}>
          Save changes
        </Button>
      </Group>
    </Form>
  )
}

const mapStateToProps = (state: State) => ({
  registration: state.ranking.registration,
})

export default connect(mapStateToProps)(EditLogForm)
