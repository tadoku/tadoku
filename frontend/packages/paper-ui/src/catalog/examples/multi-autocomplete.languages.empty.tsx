import * as stylex from '@stylexjs/stylex'
import { size } from 'paper-ui/tokens.stylex'
import { useStore } from '@tanstack/react-form'
import { Button, Flash, layout, usePaperForm } from 'paper-ui'
import { useState } from 'react'

const LANGUAGES = [
  { id: 'ja', label: 'Japanese' },
  { id: 'zh', label: 'Chinese' },
  { id: 'ko', label: 'Korean' },
]

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
    defaultValues: { multiple: [] as (typeof LANGUAGES)[number][] },
    onSubmit: ({ value: values }) => setSavedValues(values),
  })
  const values = useStore(methods.store, state => state.values)

  return (
    <methods.AppForm>
      <methods.Form xstyle={[layout.stack, styles.form]}>
        <methods.AppField
          name="multiple"
          validators={{
            onChange: ({ value }) => {
              if (!value.length) return 'Choose at least one option.'
              if (value.length !== 2)
                return 'Choose two languages for this bilingual entry.'
            },
          }}
        >
          {field => (
            <field.AutocompleteMultiInput
              label="Languages"
              hint="Choose two languages for this bilingual entry."
              required
              maxSelections={2}
              options={LANGUAGES}
              format={option => option.label}
              getId={option => option.id}
              placeholder="Add language"
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
