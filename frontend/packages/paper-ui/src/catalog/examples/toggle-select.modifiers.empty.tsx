import { Button, ToggleSelect } from 'paper-ui'
import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function Example() {
  const methods = useForm({ defaultValues: { modifiers: [] as string[] } })
  const [saved, setSaved] = useState<string | null>(null)

  return (
    <FormProvider {...methods}>
      <form
        className="paper-stack"
        noValidate
        onSubmit={methods.handleSubmit(({ modifiers }) => {
          setSaved(modifiers.length ? 'Saved with passive listening.' : 'Saved without a modifier.')
        })}
      >
        <ToggleSelect
          name="modifiers"
          label="Score modifiers"
          options={[
            { value: 'Passive listening', label: 'Passive listening', description: '×0.5' },
          ]}
        />
        <Button type="submit">Save selection</Button>
        {saved ? <p role="status">{saved}</p> : null}
      </form>
    </FormProvider>
  )
}
