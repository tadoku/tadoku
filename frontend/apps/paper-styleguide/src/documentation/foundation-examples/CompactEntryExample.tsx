import * as stylex from '@stylexjs/stylex'
import { useState } from 'react'
import { Button, densities, layout, typography, usePaperForm } from 'paper-ui'

export default function CompactEntry() {
  const methods = usePaperForm({ defaultValues: { work: 'コンビニ人間' }, onSubmit: () => setSaved(true) })
  const [saved, setSaved] = useState(false)
  return (
    <section data-density="compact" {...stylex.props(densities.compact, typography.body)}>
      <methods.AppForm>
        <methods.Form xstyle={[layout.stack]}>
          <methods.AppField name="work">{field => <field.Input label="Work" />}</methods.AppField>
          <div {...stylex.props(layout.cluster)}>
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
