import { ButtonGroup, stack, Surface, typography } from 'paper-ui'
import { useState } from 'react'

export default function ReadingLogPatternFixture() {
  const [details, setDetails] = useState(false)
  return (
    <Surface as="article" accent className={stack()}>
      <p className={typography({ role: 'metadata' })}>September 5 · Japanese</p>
      <h3 className={typography({ role: 'component' })}>コンビニ人間</h3>
      <p>48 pages read · Saved to reading history</p>
      <p>Not submitted to a contest. Notes are private.</p>
      <ButtonGroup
        actions={[
          {
            id: 'view',
            label: details ? 'Hide log details' : 'View log',
            variant: 'outline',
            onSelect: () => setDetails(!details),
          },
        ]}
      />
      {details ? (
        <section aria-label="Log details" className={stack()}>
          <h4 className={typography({ role: 'component' })}>Log details</h4>
          <p>
            Reading date: September 5, 2026. Language: Japanese. Amount: 48
            pages.
          </p>
          <p>Private note: The shop’s routines are becoming familiar.</p>
          <p>This example is local; no reading data has been saved.</p>
        </section>
      ) : null}
    </Surface>
  )
}
