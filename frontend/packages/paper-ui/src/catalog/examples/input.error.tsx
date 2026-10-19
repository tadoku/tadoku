import { usePaperForm } from 'paper-ui'
import { useEffect } from 'react'

export default function InputErrorFixture() {
  const methods = usePaperForm({ defaultValues: { pages: '' } })
  useEffect(() => {
    methods.setFieldMeta('pages', meta => ({
      ...meta,
      isTouched: true,
      errorMap: { onSubmit: 'Enter the number of pages read.' },
    }))
  }, [methods])
  return (
    <methods.AppForm>
      <methods.AppField
        name="pages"
        validators={{
          onChange: ({ value }) => {
            if (!value) return 'This field is required.'
          },
        }}
      >
        {field => (
          <field.Input
            label="Pages read"
            hint="Use whole pages."
            inputMode="numeric"
            required
          />
        )}
      </methods.AppField>
    </methods.AppForm>
  )
}
