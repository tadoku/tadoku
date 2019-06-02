import React, { FormEvent, useState } from 'react'
import { languageNameByCode, AllLanguages } from '../database'
import { connect } from 'react-redux'
import { State } from '../../store'
import { Form, Group, Label, LabelText, Select } from '../../ui/components/Form'
import { Button, StackContainer } from '../../ui/components'
import { Language } from '../interfaces'

interface Props {
  onSuccess: () => void
  onCancel: () => void
}

const sanitizeLanguageCode = (code: string) => (code === '' ? undefined : code)

const EditLogForm = ({ onSuccess: completed, onCancel: cancel }: Props) => {
  const [languages, setLanguages] = useState([
    undefined,
    undefined,
    undefined,
  ] as (string | undefined)[])

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    // @TODO: implement
  }

  return (
    <Form onSubmit={submit}>
      {languages.map((language, i) => (
        <Group key={language}>
          <Label>
            <LabelText>Language {i + 1}</LabelText>
            <Select
              value={language}
              onChange={e =>
                setLanguages(
                  languages.map((original, j) =>
                    j === i
                      ? sanitizeLanguageCode(e.currentTarget.value)
                      : original,
                  ),
                )
              }
            >
              <option value={undefined}>Choose a language...</option>
              {AllLanguages.map(l => (
                <option value={l.code}>{languageNameByCode(l.code)}</option>
              ))}
            </Select>
          </Label>
        </Group>
      ))}
      <Group>
        <StackContainer>
          <Button type="submit" primary>
            Join
          </Button>
          <Button onClick={cancel}>Nevermind, next time!</Button>
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
