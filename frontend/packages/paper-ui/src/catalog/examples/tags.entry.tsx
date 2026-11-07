import { useStore } from '@tanstack/react-form'
import { Button, cluster, Flash, stack, usePaperForm } from 'paper-ui'
import { useState } from 'react'

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = usePaperForm({
    defaultValues: { tags: ['fiction'] },
    onSubmit: ({ value: values }) => setSavedValues(values),
  })
  const values = useStore(methods.store, state => state.values)

  return (
    <methods.AppForm>
      <methods.Form className={stack({ className: "max-w-form" })}>
        <methods.AppField
          name="tags"
          validators={{
            onChange: ({ value }) => {
              if (!value.length) return 'Add at least one tag.'
            },
          }}
        >
          {field => (
            <field.TagsInput
              label="Tags"
              hint="Choose a suggestion or type a new tag and press Enter. Add up to four tags."
              required
              options={['fiction', 'history', 'manga', 'nonfiction']}
              maxSelections={4}
              placeholder="Add tag"
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
