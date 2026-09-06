import { Input } from 'paper-ui'
import { useEffect } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function InputErrorFixture() {
  const methods = useForm<{ pages: string }>({ defaultValues: { pages: '' } })
  useEffect(() => {
    methods.setError('pages', { message: 'Enter the number of pages read.' })
  }, [methods])
  return (
    <FormProvider {...methods}>
      <Input
        name="pages"
        label="Pages read"
        hint="Use whole pages."
        inputMode="numeric"
        required
      />
    </FormProvider>
  )
}
