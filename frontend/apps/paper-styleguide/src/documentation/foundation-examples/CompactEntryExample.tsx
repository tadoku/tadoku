import { usePaperForm, Button } from 'paper-ui'
import { useState } from 'react'

export default function CompactEntry() {
  const methods = usePaperForm({ defaultValues: { work: 'コンビニ人間' }, onSubmit: () => setSaved(true) })
  const [saved, setSaved] = useState(false)
  return (
    <section data-density="compact" className="paper-type-body">
      <methods.AppForm>
        <methods.Form className="paper-stack">
          <methods.AppField name="work">{field => <field.Input label="Work" />}</methods.AppField>
          <div className="paper-cluster">
            <Button type="submit">Save reading</Button>
          </div>
          <p role="status">
            {saved ? 'Saved in this demo.' : 'No entry saved.'}
          </p>
        </methods.Form>
      </methods.AppForm>
    </section>
  )
}
