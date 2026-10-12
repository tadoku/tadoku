import { EmptyState, SectionHeading, Surface, buttonClassName } from 'paper-ui'

export default function Example() {
  return (
    <Surface as="section" aria-labelledby="year-heading">
      <SectionHeading level={2} id="year-heading" title="Anton’s activity" />
      <EmptyState
        variant="plain"
        headingLevel={3}
        title="No activity recorded in 2026"
        description="A few pages or minutes are a good place to start. Your personal record grows with every entry, whether you join a contest or read on your own."
        actions={
          <a className={buttonClassName()} href="#log-activity">
            Log activity
          </a>
        }
      />
    </Surface>
  )
}
