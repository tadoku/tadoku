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
    defaultValues: { notes: 'Finished chapter two. The pacing is picking up.' },
    onSubmit: ({ value: values }) => setSavedValues(values),
  })
  const values = useStore(methods.store, state => state.values)

  return (
    <methods.AppForm>
      <methods.Form xstyle={[layout.stack, styles.form]}>
        <methods.AppField
          name="notes"
          validators={{
            onChange: ({ value }) => {
              if (!value) return 'This field is required.'
            },
          }}
        >
          {field => (
            <field.TextArea
              label="Reading notes"
              hint="Keep spoilers out of public notes."
              required
            />
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
