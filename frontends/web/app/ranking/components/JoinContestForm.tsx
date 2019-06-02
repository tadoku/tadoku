import React, { FormEvent, useState } from 'react'
import { languageNameByCode, AllLanguages } from '../database'
import { connect } from 'react-redux'
import { State } from '../../store'
import { Form, Group, Label, Select } from '../../ui/components/Form'
import { Button, StackContainer } from '../../ui/components'

interface Props {
  onSuccess: () => void
  onCancel: () => void
}

const sanitizeLanguageCode = (code: string) => (code === '' ? undefined : code)

// @TODO: extract out
const maxLanguages = 3

const EditLogForm = ({ onSuccess: completed, onCancel: cancel }: Props) => {
  const [languages, setLanguages] = useState([undefined] as (
    | string
    | undefined)[])

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    // @TODO: implement
    if (false) {
      completed()
    }
  }

  const addLanguage = () => {
    if (languages.length >= maxLanguages) {
      return
    }

    setLanguages([...languages, undefined])
  }

  return (
    <Form onSubmit={submit}>
      <Group>
        {languages.map((language, i) => (
          <Label key={language}>
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
        ))}

        {languages.length < maxLanguages && (
          <Button onClick={addLanguage} icon="plus" plain small>
            Add language
          </Button>
        )}
      </Group>
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
