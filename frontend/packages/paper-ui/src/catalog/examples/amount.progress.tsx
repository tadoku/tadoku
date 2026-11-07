import * as stylex from '@stylexjs/stylex'
import { size } from 'paper-ui/tokens.stylex'
import { useStore } from '@tanstack/react-form'
import { Button, Flash, layout, usePaperForm } from 'paper-ui'
import { useState } from 'react'

const styles = stylex.create({
  form: { maxInlineSize: size[512] },
  values: { margin: 0, overflowWrap: 'anywhere', whiteSpace: 'pre-wrap' },
})

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = usePaperForm({
    defaultValues: { progressValue: 48, progressUnit: 'pages' },
    onSubmit: ({ value: values }) => setSavedValues(values),
  })
  const values = useStore(methods.store, state => state.values)

  return (
    <methods.AppForm>
      <methods.Form xstyle={[layout.stack, styles.form]}>
        <methods.AppField name="progressUnit">
          {unitField => (
            <methods.AppField
              name="progressValue"
              validators={{
                onChange: ({ value }) => {
                  if (!Number.isFinite(value)) return 'Enter an amount.'
                  if (value < 1) return 'Enter at least 1.'
                },
              }}
            >
              {field => (
                <field.AmountWithUnit
                  label="Progress"
                  hint="Enter a positive amount; choose pages or minutes."
                  min={1}
                  required
                  units={[
                    { value: 'pages', label: 'pages' },
                    { value: 'minutes', label: 'minutes' },
                  ]}
                  unitField={unitField}
                />
              )}
            </methods.AppField>
          )}
        </methods.AppField>
        <div {...stylex.props(layout.cluster)}>
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
              {...stylex.props(styles.values)}
            >
              {JSON.stringify(savedValues, null, 2)}
            </pre>
          </Flash>
        ) : null}
        <details>
          <summary>Current form values</summary>
          <pre {...stylex.props(styles.values)}>
            {JSON.stringify(values, null, 2)}
          </pre>
        </details>
      </methods.Form>
    </methods.AppForm>
  )
}
