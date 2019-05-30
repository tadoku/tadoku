import React, { FormEvent, useState, useEffect } from 'react'
import styled from 'styled-components'
import Constants from '../../ui/Constants'
import { AllMediums, languageNameByCode } from '../database'
import { connect } from 'react-redux'
import { State } from '../../store'
import { RankingRegistration, ContestLog } from '../interfaces'
import RankingApi from '../api'
import Router from 'next/router'

const Form = styled.form``
const Label = styled.label`
  display: block;
  margin-bottom: 10px;
`
const LabelText = styled.span`
  display: block;
`
const Input = styled.input`
  border: none;
  background: ${Constants.colors.secondary};
  padding: 8px 20px;
  font-size: 1.1em;
  height: 36px;
`

const Button = styled.button`
  border: none;
  background: ${Constants.colors.secondary};
  padding: 8px 20px;
  font-size: 1.1em;
  height: 36px;
`

interface Props {
  log: ContestLog
  registration: RankingRegistration | undefined
}

const EditLogForm = ({ log, registration }: Props) => {
  const [amount, setAmount] = useState(log.amount.toString())
  const [mediumId, setMediumId] = useState(log.mediumId.toString())
  const [languageCode, setLanguageCode] = useState(log.languageCode)

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    // const success = await RankingApi.updateLog(log.id, {
    //   contestId: log.contestId,
    //   mediumId: Number(mediumId),
    //   amount: Number(amount),
    //   languageCode,
    // })

    // if (success) {
    //   // @TODO
    // }
  }

  if (!registration) {
    return null
  }

  return (
    <Form onSubmit={submit}>
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

      <LabelText>Language</LabelText>
      {registration.languages.map(code => (
        <Label key={code}>
          <input
            type="radio"
            value={code}
            checked={code === languageCode}
            onChange={e => setLanguageCode(e.target.value)}
          />
          {languageNameByCode(code)}
        </Label>
      ))}
      <Button type="submit">Submit pages</Button>
    </Form>
  )
}

const mapStateToProps = (state: State) => ({
  registration: state.ranking.registration,
})

export default connect(mapStateToProps)(EditLogForm)
