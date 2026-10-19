import { Button, usePaperForm } from 'paper-ui'
import { useState } from 'react'

export default function Example() {
  const methods = usePaperForm({
    defaultValues: { modifiers: [] as string[] },
    onSubmit: ({ value: { modifiers } }) => {
      setSaved(
        modifiers.length
          ? 'Saved with passive listening.'
          : 'Saved without a modifier.',
      )
    },
  })
  const [saved, setSaved] = useState<string | null>(null)

  return (
    <methods.AppForm>
      <methods.Form className="paper-stack">
        <methods.AppField name="modifiers">
          {field => (
            <field.ToggleSelect
              label="Score modifiers"
              options={[
                {
                  value: 'Passive listening',
                  label: 'Passive listening',
                  description: '×0.5',
                },
              ]}
            />
          )}
        </methods.AppField>
        <Button type="submit">Save selection</Button>
        {saved ? <p role="status">{saved}</p> : null}
      </methods.Form>
    </methods.AppForm>
  )
}
