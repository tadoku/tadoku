import { AmountWithUnit, Button, Flash } from 'paper-ui'
import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = useForm({
    defaultValues: { progressValue: 48, progressUnit: 'pages' },
  })
  return (
    <FormProvider {...methods}>
      <form
        className="paper-stack"
        style={{ maxWidth: '32rem' }}
        noValidate
        onSubmit={methods.handleSubmit(values => setSavedValues(values))}
      >
        <AmountWithUnit
          name="progress"
          label="Progress"
          hint="Enter a positive amount; choose pages or minutes."
          min={1}
          required
          units={[
            { value: 'pages', label: 'pages' },
            { value: 'minutes', label: 'minutes' },
          ]}
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
