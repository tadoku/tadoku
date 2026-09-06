import { Input } from 'paper-ui'
import { FormProvider, useForm } from 'react-hook-form'

export default function InputStatesFixture() {
  const methods = useForm({
    defaultValues: { readonly: 'Japanese', disabled: 'Archived' },
  })
  return (
    <FormProvider {...methods}>
      <div className="paper-stack">
        <Input name="readonly" label="Language" readOnly />
        <Input name="disabled" label="Status" disabled />
      </div>
    </FormProvider>
  )
}
