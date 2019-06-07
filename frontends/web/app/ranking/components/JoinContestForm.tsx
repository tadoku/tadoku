import React, { FormEvent, useState } from 'react'
import { languageNameByCode, AllLanguages } from '../database'
import { connect } from 'react-redux'
import { State } from '../../store'
import { Form, Group, Label, Select } from '../../ui/components/Form'
import { Button, StackContainer } from '../../ui/components'
import RankingApi from '../api'
import { Contest } from '../../contest/interfaces'
import { validateLanguageCode } from '../domain'

interface Props {
  contest: Contest
  onSuccess: () => void
  onCancel: () => void
}

const sanitizeLanguageCode = (code: string) => (code === '' ? undefined : code)

// @TODO: extract out
const maxLanguages = 3

const JoinContestForm = ({
  contest,
  onSuccess: completed,
  onCancel: cancel,
}: Props) => {
  const [languages, setLanguages] = useState([undefined] as (
    | string
    | undefined)[])

  const submit = async (event: FormEvent) => {
    event.preventDefault()

    const languageCodes = languages.filter(l => !!l) as string[]
    const success = await RankingApi.joinContest(contest.id, languageCodes)

    if (success) {
      completed()
    }
  }

  const addLanguage = () => {
    if (languages.length >= maxLanguages) {
      return
    }

    setLanguages([...languages, undefined])
  }

  const validate = () => validateLanguageCode(languages[0] || '')

  const hasError = {
    form: !validate(),
  }

  return (
    <Form onSubmit={submit}>
      <Group>
        {languages.map((language, i) => (
          <Label key={i}>
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
              <option value="">Choose a language...</option>
              {AllLanguages.map(l => (
                <option value={l.code} key={l.code}>
                  {languageNameByCode(l.code)}
                </option>
              ))}
            </Select>
          </Label>
        ))}

        {languages.length < maxLanguages && (
          <Button type="button" onClick={addLanguage} icon="plus" plain small>
            Add language
          </Button>
        )}
      </Group>
      <Group>
        <StackContainer>
          <Button type="submit" primary disabled={hasError.form}>
            Join
          </Button>
          <Button type="button" onClick={cancel}>
            Nevermind, next time!
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

export default connect(mapStateToProps)(JoinContestForm)
