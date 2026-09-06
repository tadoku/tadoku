import { RadioSelect } from 'paper-ui'
import { FormProvider, useForm } from 'react-hook-form'

export default function DefaultRadioFixture() {
  const methods = useForm({ defaultValues: { viewport: 'tablet' } })
  return (
    <FormProvider {...methods}>
      <RadioSelect
        name="viewport"
        label="Preview size"
        options={[
          { value: 'phone', label: 'Phone' },
          { value: 'tablet', label: 'Tablet' },
          { value: 'desktop', label: 'Desktop', disabled: true },
        ]}
        hint="Desktop preview is unavailable in this example."
      />
    </FormProvider>
  )
}
