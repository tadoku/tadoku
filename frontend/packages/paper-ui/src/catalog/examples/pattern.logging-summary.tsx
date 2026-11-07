import * as stylex from '@stylexjs/stylex'
import { ButtonGroup, Surface, layout, typography } from 'paper-ui'
import { useState } from 'react'

export default function ReadingLogPatternFixture() {
  const [details, setDetails] = useState(false)
  return (
    <Surface as="article" accent {...stylex.props(layout.stack)}>
      <p {...stylex.props(typography.metadata)}>September 5 · Japanese</p>
      <h3 {...stylex.props(typography.component)}>コンビニ人間</h3>
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
        <section aria-label="Log details" {...stylex.props(layout.stack)}>
          <h4 {...stylex.props(typography.component)}>Log details</h4>
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
