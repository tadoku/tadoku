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
  RadioButton,
  GroupError,
} from '../../ui/components/Form'
import { Button, StackContainer } from '../../ui/components'
import { validateLanguageCode, validateAmount } from '../domain'

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

    return '1'
  })
  const [languageCode, setLanguageCode] = useState(() => {
    if (log) {
      return log.languageCode
    }

    return ''
  })

  if (!registration) {
    return null
  }

  if (languageCode === '') {
    setLanguageCode(registration.languages[0])
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

  const validate = () =>
    validateAmount(amount) && validateLanguageCode(languageCode)

  const hasError = {
    form: !validate(),
    amount: amount !== '' && !validateAmount(amount),
    languageCode: languageCode != '' && !validateLanguageCode(languageCode),
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
            error={hasError.amount}
          />
          <GroupError message="Invalid page count" hidden={!hasError.amount} />
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
          <RadioButton
            key={code}
            type="radio"
            value={code}
            checked={code === languageCode}
            onChange={e => setLanguageCode(e.target.value)}
            label={languageNameByCode(code)}
          />
        ))}
        <GroupError
          message="Invalid language"
          hidden={!hasError.languageCode}
        />
      </Group>
      <Group>
        <StackContainer>
          <Button type="submit" disabled={hasError.form} primary>
            Save changes
          </Button>
          <Button type="button" onClick={cancel}>
            Cancel
          </Button>
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
