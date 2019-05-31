import React, { FormEvent, useState, useEffect } from 'react'
import { AllMediums, languageNameByCode } from '../database'
import { connect } from 'react-redux'
import { State } from '../../store'
import { RankingRegistration } from '../interfaces'
import RankingApi from '../api'
import Router from 'next/router'
import { Form, Label, LabelText, Input } from '../../ui/components/Form'
import { Button } from '../../ui/components'

interface Props {
  registration: RankingRegistration | undefined
}

const UpdateForm = ({ registration }: Props) => {
  const [amount, setAmount] = useState('')
  const [mediumId, setMediumId] = useState('1')
  const [languageCode, setLanguageCode] = useState('')

  useEffect(() => {
    if (registration) {
      setLanguageCode(registration.languages[0])
    }
  }, [registration])

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    // TODO: add validation
    if (!registration) {
      return
    }

    const success = await RankingApi.createLog({
      contestId: registration.contestId,
      mediumId: Number(mediumId),
      amount: Number(amount),
      languageCode,
    })

    if (success) {
      Router.push('/')
    }
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

export default connect(mapStateToProps)(UpdateForm)
