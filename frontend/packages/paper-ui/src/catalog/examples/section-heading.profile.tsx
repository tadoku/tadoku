import { SectionHeading, buttonClassName } from 'paper-ui'

export default function Example() {
  return (
    <section aria-labelledby="activity-heading">
      <SectionHeading
        level={2}
        id="activity-heading"
        title="Anton’s activity"
        description="Reading and listening recorded in 2026."
        actions={
          <a className={buttonClassName({ variant: 'outline' })} href="#activity-log">
            View activity log
          </a>
        }
      />
      <p>1,240.5 personal points across 18 entries.</p>
      <section aria-labelledby="rhythm-heading">
        <SectionHeading
          level={3}
          id="rhythm-heading"
          title="Immersion rhythm"
          actions={
            <a className={buttonClassName({ variant: 'link' })} href="#scoring">
              How scoring works
            </a>
          }
        />
        <p>12 active days · Longest streak: 5 days</p>
      </section>
    </section>
  )
}
