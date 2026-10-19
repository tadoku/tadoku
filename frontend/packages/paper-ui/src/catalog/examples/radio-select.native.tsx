import { usePaperForm } from 'paper-ui'

export default function DefaultRadioFixture() {
  const methods = usePaperForm({ defaultValues: { viewport: 'tablet' } })
  return (
    <methods.AppForm>
      <methods.AppField name="viewport">
        {field => (
          <field.RadioSelect
            label="Preview size"
            options={[
              { value: 'phone', label: 'Phone' },
              { value: 'tablet', label: 'Tablet' },
              { value: 'desktop', label: 'Desktop', disabled: true },
            ]}
            hint="Desktop preview is unavailable in this example."
          />
        )}
      </methods.AppField>
    </methods.AppForm>
  )
}
