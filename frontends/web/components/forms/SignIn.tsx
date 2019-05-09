import React, { FormEvent, useState } from 'react'
import styled from 'styled-components'
import Constants from '../ui/Constants'
import { SignIn } from '../../domain/Api'

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

const SignInForm = () => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const response = await SignIn({ email, password })
    console.log(response)
  }

  return (
    <Form onSubmit={submit}>
      <Label>
        <LabelText>Email</LabelText>
        <Input
          type="email"
          placeholder="tadoku@example.com"
          value={email}
          onChange={e => setEmail(e.target.value)}
        />
      </Label>
      <Label>
        <LabelText>Password</LabelText>
        <Input
          type="password"
          value={password}
          onChange={e => setPassword(e.target.value)}
        />
      </Label>
      <Button type="submit">Sign in</Button>
    </Form>
  )
}

export default SignInForm
