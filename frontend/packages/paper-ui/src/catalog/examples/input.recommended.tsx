import { Button, usePaperForm } from 'paper-ui'
import { useState } from 'react'

export default function InputRecommendedFixture() {
  const [saved, setSaved] = useState('')
  const methods = usePaperForm({
    defaultValues: ({ title: 'August reading log', date: '2026-08-31', pages: 24 } as { title: string; date: string; pages: number }), onSubmit: ({ value: { title, date, pages } }) => setSaved(`Saved: ${title} · ${pages} pages · ${date}`) })
  return (
    <methods.AppForm>
      <methods.Form className="paper-stack w-full max-w-2xl">
        <methods.AppField name="title" validators={{ onChange: ({ value }) => { if (!value) return ('Enter a log title.'); } }}>{field => <field.Input label="Log title" required />}</methods.AppField>
        <div className="grid gap-4 sm:grid-cols-2">
          <methods.AppField name="pages" validators={{ onChange: ({ value }) => { if (String(value) === '' || !Number.isFinite(Number(value))) return ('Enter the number of pages.'); if (Number(value) < 1) return 'Enter at least one page.'; } }}>{field => <field.Input label="Pages read" type="number" min={1} required />}</methods.AppField>
          <methods.AppField name="date" validators={{ onChange: ({ value }) => { if (!value) return ('Choose a date.'); } }}>{field => <field.Input label="Date" type="date" hint="The date you read, in UTC. Fields stay aligned when only one has a hint." required />}</methods.AppField>
        </div>
        <div className="paper-cluster">
          <Button type="submit">Save log</Button>
        </div>
        {saved && <p role="status">{saved}</p>}
      </methods.Form>
    </methods.AppForm>
  )
}
