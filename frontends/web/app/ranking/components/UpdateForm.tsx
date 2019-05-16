import React, { FormEvent, useState } from 'react'
import styled from 'styled-components'
import Constants from '../../ui/Constants'

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

const UpdateForm = () => {
  const [amount, setAmount] = useState('' as string)

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
          step={1}
        />
      </Label>
      <Button type="submit">Submit pages</Button>
    </Form>
  )
}

export default UpdateForm
