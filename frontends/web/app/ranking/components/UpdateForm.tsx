import React, { FormEvent, useState, useEffect } from 'react'
import styled from 'styled-components'
import Constants from '../../ui/Constants'
import { AllMediums } from '../database'
import { connect } from 'react-redux'
import { State } from '../../store'
import { RankingRegistration } from '../interfaces'
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
  registration: RankingRegistration | undefined
}

const UpdateForm = ({ registration }: Props) => {
  const [amount, setAmount] = useState('')
  const [medium, setMedium] = useState('0')

  useEffect(() => {
    if (!registration) {
      Router.push('/')
    }
  }, [])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
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
        <select value={medium} onChange={e => setMedium(e.target.value)}>
          {AllMediums.map(m => (
            <option value={m.id}>{m.description}</option>
          ))}
        </select>
      </Label>
      <Button type="submit">Submit pages</Button>
    </Form>
  )
}

const mapStateToProps = (state: State) => ({
  registration: state.ranking.registration,
})

export default connect(mapStateToProps)(UpdateForm)
