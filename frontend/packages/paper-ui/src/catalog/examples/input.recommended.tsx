import * as stylex from '@stylexjs/stylex'
import { Button, layout, usePaperForm } from 'paper-ui'
import { breakpoints, size, space } from 'paper-ui/tokens.stylex'
import { useState } from 'react'

const styles = stylex.create({
  form: { inlineSize: '100%', maxInlineSize: size[672] },
  pair: {
    gap: space[16],
    display: 'grid',
    gridTemplateColumns: { default: '1fr', [breakpoints.wide]: 'repeat(2, minmax(0, 1fr))' },
  },
})

export default function InputRecommendedFixture() {
  const [saved, setSaved] = useState('')
  const methods = usePaperForm({
    defaultValues: {
      title: 'August reading log',
      date: '2026-08-31',
      pages: 24,
    },
    onSubmit: ({ value: { title, date, pages } }) =>
      setSaved(`Saved: ${title} · ${pages} pages · ${date}`),
  })
  return (
    <methods.AppForm>
      <methods.Form xstyle={[layout.stack, styles.form]}>
        <methods.AppField
          name="title"
          validators={{
            onChange: ({ value }) => {
              if (!value) return 'Enter a log title.'
            },
          }}
        >
          {field => <field.Input label="Log title" required />}
        </methods.AppField>
        <div {...stylex.props(styles.pair)}>
          <methods.AppField
            name="pages"
            validators={{
              onChange: ({ value }) => {
                if (!Number.isFinite(value)) return 'Enter the number of pages.'
                if (value < 1) return 'Enter at least one page.'
              },
            }}
          >
            {field => (
              <field.Input label="Pages read" type="number" min={1} required />
            )}
          </methods.AppField>
          <methods.AppField
            name="date"
            validators={{
              onChange: ({ value }) => {
                if (!value) return 'Choose a date.'
              },
            }}
          >
            {field => (
              <field.Input
                label="Date"
                type="date"
                hint="The date you read, in UTC. Fields stay aligned when only one has a hint."
                required
              />
            )}
          </methods.AppField>
        </div>
        <div {...stylex.props(layout.cluster)}>
          <Button type="submit">Save log</Button>
        </div>
        {saved && <p role="status">{saved}</p>}
      </methods.Form>
    </methods.AppForm>
  )
}
