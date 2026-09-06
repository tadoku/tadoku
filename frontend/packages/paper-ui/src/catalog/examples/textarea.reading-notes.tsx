import { Button, Flash, TextArea } from 'paper-ui'
import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = useForm({
    defaultValues: { notes: 'Finished chapter two. The pacing is picking up.' },
  })
  return (
    <FormProvider {...methods}>
      <form
        className="paper-stack"
        style={{ maxWidth: '32rem' }}
        noValidate
        onSubmit={methods.handleSubmit(values => setSavedValues(values))}
      >
        <TextArea
          name="notes"
          label="Reading notes"
          hint="Keep spoilers out of public notes."
          required
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
