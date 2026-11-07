import { Button, Surface, typography } from 'paper-ui'

export default function ReadingSection() {
  return (
    <section className="mx-auto grid w-full max-w-content-wide grid-cols-1 gap-6 lg:grid-cols-two-one">
      <Surface className="min-w-0 flex flex-col gap-4 p-4 md:p-6">
        <h2 className={typography({ role: 'section' , className: 'm-0' })}>September reading</h2>
        <p className="m-0 max-w-measure">Choose a reading session to continue.</p>
        <div className="flex flex-wrap items-center gap-inline">
          <Button variant="outline">All languages</Button>
          <Button variant="outline">September 2026</Button>
          <Button>Log reading</Button>
        </div>
      </Surface>
      <aside className="min-w-0 flex flex-col gap-2 py-4">
        <h2 className={typography({ role: 'component' , className: 'm-0' })}>Your reading goal</h2>
        <p className="m-0">Read for 20 minutes on four days this week.</p>
        <p className="m-0 text-muted">Three sessions complete. One more to go.</p>
      </aside>
    </section>
  )
}
