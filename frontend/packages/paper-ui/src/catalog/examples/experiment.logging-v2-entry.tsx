import { Button, Flash, Input, Surface } from 'paper-ui'
import { useEffect, useRef, useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function LoggingExperimentFixture() {
  const methods = useForm({
    defaultValues: { title: 'コンビニ人間', pages: '48' },
  })
  const [reviewing, setReviewing] = useState(false)
  const heading = useRef<HTMLHeadingElement>(null)
  useEffect(() => {
    if (reviewing) heading.current?.focus()
    else if (methods.formState.isSubmitted) methods.setFocus('title')
  }, [reviewing, methods])
  return (
    <FormProvider {...methods}>
      {reviewing ? (
        <Surface className="paper-stack">
          <h3 ref={heading} tabIndex={-1} className="paper-type-component">
            Review reading
          </h3>
          <p>
            {methods.getValues('pages')} pages of {methods.getValues('title')}
          </p>
          <Flash title="Preview only">
            Nothing has been saved or submitted to a contest.
          </Flash>
          <div>
            <Button variant="outline" onClick={() => setReviewing(false)}>
              Edit entry
            </Button>
          </div>
        </Surface>
      ) : (
        <form
          className="paper-stack"
          onSubmit={methods.handleSubmit(() => setReviewing(true))}
        >
          <Flash title="Experimental flow">
            Review the entry before considering a separate contest step.
          </Flash>
          <Input
            name="title"
            label="Work"
            required
            rules={{
              validate: value =>
                String(value).trim().length > 0 || 'Enter a work title.',
            }}
          />
          <Input
            name="pages"
            label="Pages"
            type="number"
            inputMode="numeric"
            min={1}
            step={1}
            required
            rules={{ min: { value: 1, message: 'Enter at least one page.' } }}
          />
          <div className="paper-cluster">
            <Button type="submit">Review entry</Button>
          </div>
        </form>
      )}
    </FormProvider>
  )
}
