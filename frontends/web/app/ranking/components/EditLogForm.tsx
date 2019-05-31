import React, { FormEvent, useState } from 'react'
import styled from 'styled-components'
import Constants from '../../ui/Constants'
import { AllMediums, languageNameByCode } from '../database'
import { connect } from 'react-redux'
import { State } from '../../store'
import { RankingRegistration, ContestLog } from '../interfaces'
import RankingApi from '../api'

const Form = styled.form``

const Group = styled.div`
  & + & {
    margin-top: 30px;
  }
`

const Label = styled.label`
  display: block;
`
const LabelText = styled.span`
  display: block;
  font-weight: 600;
  font-size: 1.3em;
  margin-bottom: 7px;
`

const LabelForRadio = styled(Label)`
  padding: 3px 0;
  line-height: 1em;

  span {
    margin-left: 5px;
  }

  input:checked + span {
    font-weight: 600;
  }
`

const Input = styled.input`
  border: none;
  background: ${Constants.colors.secondary};
  padding: 8px 20px;
  font-size: 1.1em;
  height: 36px;
  display: block;
  width: 100%;
  box-sizing: border-box;
`

const Button = styled.button`
  border: none;
  background: ${Constants.colors.secondary};
  padding: 8px 20px;
  font-size: 1.1em;
  height: 36px;
  width: 100%;
  box-sizing: border-box;
`

interface Props {
  log: ContestLog
  registration: RankingRegistration | undefined
  onSuccess: () => void
}

const EditLogForm = ({ log, registration, onSuccess: completed }: Props) => {
  const [amount, setAmount] = useState(log.amount.toString())
  const [mediumId, setMediumId] = useState(log.mediumId.toString())
  const [languageCode, setLanguageCode] = useState(log.languageCode)

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
          <select value={mediumId} onChange={e => setMediumId(e.target.value)}>
            {AllMediums.map(m => (
              <option value={m.id} key={m.id}>
                {m.description}
              </option>
            ))}
          </select>
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
        <Button type="submit">Submit pages</Button>
      </Group>
    </Form>
  )
}

const mapStateToProps = (state: State) => ({
  registration: state.ranking.registration,
})

export default connect(mapStateToProps)(EditLogForm)
