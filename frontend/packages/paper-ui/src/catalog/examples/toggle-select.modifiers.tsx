import { Button, ToggleSelect } from 'paper-ui'
import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function Example() {
  const methods = useForm({ defaultValues: { modifiers: ['Manga'] } })
  const [saved, setSaved] = useState<string | null>(null)

  return (
    <FormProvider {...methods}>
      <form
        className="paper-stack"
        noValidate
        onSubmit={methods.handleSubmit(({ modifiers }) => {
          setSaved(modifiers.length ? `Saved: ${modifiers[0]}.` : 'Saved without a modifier.')
        })}
      >
        <ToggleSelect
          name="modifiers"
          label="Score modifiers"
          hint="Choose one, or leave all off."
          options={[
            { value: 'Manga', label: 'Manga', description: '×0.2' },
            { value: 'Comic', label: 'Comic', description: '×0.2' },
            { value: 'Two column', label: 'Two column', description: '×1.6' },
          ]}
        />
        <div className="paper-cluster">
          <Button type="submit">Save selection</Button>
          <Button variant="ghost" onClick={() => { methods.reset(); setSaved(null) }}>
            Reset
          </Button>
        </div>
        {saved ? <p role="status">{saved}</p> : null}
      </form>
    </FormProvider>
  )
}
