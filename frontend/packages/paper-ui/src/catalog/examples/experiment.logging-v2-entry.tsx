import { Button, Flash, Surface, usePaperForm } from 'paper-ui'
import { useEffect, useRef, useState } from 'react'

export default function LoggingExperimentFixture() {
  const methods = usePaperForm({
    defaultValues: { title: 'コンビニ人間', pages: 48 }, onSubmit: () => setReviewing(true) })
  const [reviewing, setReviewing] = useState(false)
  const heading = useRef<HTMLHeadingElement>(null)
  useEffect(() => {
    if (reviewing) heading.current?.focus()
    else if (methods.state.submissionAttempts > 0) document.getElementsByName('title')[0]?.focus()
  }, [reviewing, methods])
  return (
    <methods.AppForm>
      {reviewing ? (
        <Surface className="paper-stack">
          <h3 ref={heading} tabIndex={-1} className="paper-type-component">
            Review reading
          </h3>
          <p>
            {methods.getFieldValue('pages')} pages of {methods.getFieldValue('title')}
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
        <methods.Form className="paper-stack">
          <Flash title="Experimental flow">
            Review the entry before considering a separate contest step.
          </Flash>
          <methods.AppField name="title" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; const result = (value =>
                String(value).trim().length > 0 || 'Enter a work title.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.Input label="Work" required />}</methods.AppField>
          <methods.AppField name="pages" validators={{ onChange: ({ value }) => { if (String(value) === '' || !Number.isFinite(Number(value))) return "This field is required."; if (Number(value) < 1) return 'Enter at least one page.'; } }}>{field => <field.Input label="Pages" type="number" inputMode="numeric" min={1} step={1} required />}</methods.AppField>
          <div className="paper-cluster">
            <Button type="submit">Review entry</Button>
          </div>
        </methods.Form>
      )}
    </methods.AppForm>
  )
}
