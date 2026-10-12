import { DescriptionList, SectionHeading } from 'paper-ui'

export default function Example() {
  return (
    <div className="paper-stack" style={{ maxWidth: '36rem' }}>
      <section aria-labelledby="dates-heading">
        <SectionHeading level={3} id="dates-heading" title="Dates & registration" />
        <DescriptionList
          items={[
            { label: 'Round begins', value: 'Sep 1, 2026, 12:00 AM UTC' },
            { label: 'Round ends', value: 'Sep 14, 2026, 11:59 PM UTC' },
            { label: 'Joining this round', value: 'Open until Sep 7, 2026, 11:59 PM UTC' },
          ]}
        />
      </section>
      <section aria-labelledby="current-heading">
        <SectionHeading level={3} id="current-heading" title="Current contest" />
        <DescriptionList
          layout="inline"
          items={[
            { label: 'Period', value: 'Sep 1 – 14, 2026' },
            { label: 'Registration deadline', value: 'Sep 7, 2026' },
            { label: 'Submitted entries', value: '42' },
          ]}
        />
      </section>
    </div>
  )
}
