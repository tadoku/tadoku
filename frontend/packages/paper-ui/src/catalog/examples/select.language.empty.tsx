import { Button, Flash, Select } from 'paper-ui'
import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

const LANGUAGES = [
  { id: 'ja', label: 'Japanese' },
  { id: 'zh', label: 'Chinese' },
  { id: 'ko', label: 'Korean' },
]

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = useForm({ defaultValues: { language: '' } })
  return (
    <FormProvider {...methods}>
      <form
        className="paper-stack"
        style={{ maxWidth: '32rem' }}
        noValidate
        onSubmit={methods.handleSubmit(values => setSavedValues(values))}
      >
        <Select
          name="language"
          label="Language"
          hint="Choose the language you read."
          required
          placeholder="Choose a language"
          options={LANGUAGES.map(({ id, label }) => ({ value: id, label }))}
        />
        <div className="paper-cluster">
          <Button type="submit">Save entry</Button>
          <Button
            variant="outline"
            onClick={() => {
              methods.reset()
              setSavedValues(null)
            }}
          >
            Reset
          </Button>
        </div>
        {savedValues ? (
          <Flash variant="success" title="Entry saved">
            <pre
              style={{
                whiteSpace: 'pre-wrap',
                overflowWrap: 'anywhere',
                margin: 0,
              }}
            >
              {JSON.stringify(savedValues, null, 2)}
            </pre>
          </Flash>
        ) : null}
        <details>
          <summary>Current form values</summary>
          <pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>
            {JSON.stringify(methods.watch(), null, 2)}
          </pre>
        </details>
      </form>
    </FormProvider>
  )
}
