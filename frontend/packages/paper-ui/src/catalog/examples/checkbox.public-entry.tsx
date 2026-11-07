import { useStore } from '@tanstack/react-form'
import { Button, cluster, Flash, stack, usePaperForm } from 'paper-ui'
import { useState } from 'react'

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = usePaperForm({
    defaultValues: { public: true },
    onSubmit: ({ value: values }) => setSavedValues(values),
  })
  const values = useStore(methods.store, state => state.values)

  return (
    <methods.AppForm>
      <methods.Form className={stack({ className: "max-w-form" })}>
        <methods.AppField name="public">
          {field => (
            <field.Checkbox
              label="Show this entry on my profile"
              hint="Contest moderators can always review submitted entries."
            />
          )}
        </methods.AppField>
        <div className={cluster()}>
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
            <pre className="m-0 whitespace-pre-wrap wrap-anywhere">
              {JSON.stringify(savedValues, null, 2)}
            </pre>
          </Flash>
        ) : null}
        <details>
          <summary>Current form values</summary>
          <pre className="whitespace-pre-wrap wrap-anywhere">
            {JSON.stringify(values, null, 2)}
          </pre>
        </details>
      </methods.Form>
    </methods.AppForm>
  )
}
