import { Button, usePaperForm } from 'paper-ui'
import { useState } from 'react'

export default function Example() {
  const methods = usePaperForm({
    defaultValues: { modifiers: ['Manga'] },
    onSubmit: ({ value: { modifiers } }) => {
      setSaved(
        modifiers.length
          ? `Saved: ${modifiers[0]}.`
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
              hint="Choose one, or leave all off."
              options={[
                { value: 'Manga', label: 'Manga', description: '×0.2' },
                { value: 'Comic', label: 'Comic', description: '×0.2' },
                {
                  value: 'Two column',
                  label: 'Two column',
                  description: '×1.6',
                },
              ]}
            />
          )}
        </methods.AppField>
        <div className="paper-cluster">
          <Button type="submit">Save selection</Button>
          <Button
            variant="ghost"
            onClick={() => {
              methods.reset()
              setSaved(null)
            }}
          >
            Reset
          </Button>
        </div>
        {saved ? <p role="status">{saved}</p> : null}
      </methods.Form>
    </methods.AppForm>
  )
}
