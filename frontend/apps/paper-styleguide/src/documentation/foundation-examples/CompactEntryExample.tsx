import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import { Button, Input } from 'paper-ui'

export default function CompactEntry() {
  const methods = useForm({ defaultValues: { work: 'コンビニ人間' } })
  const [saved, setSaved] = useState(false)
  return (
    <section data-density="compact" className="paper-type-body">
      <FormProvider {...methods}>
        <form
          className="paper-stack"
          onSubmit={methods.handleSubmit(() => setSaved(true))}
        >
          <Input name="work" label="Work" />
          <div className="paper-cluster">
            <Button type="submit">Save reading</Button>
          </div>
          <p role="status">
            {saved ? 'Saved in this demo.' : 'No entry saved.'}
          </p>
        </form>
      </FormProvider>
    </section>
  )
}
