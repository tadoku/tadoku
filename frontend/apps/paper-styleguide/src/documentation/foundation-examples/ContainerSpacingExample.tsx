import { Surface, typography } from 'paper-ui'

export default function ReadingSummary() {
  return (
    <Surface className="flex flex-col gap-4 p-6">
      <h2 className={typography({ role: 'component' , className: 'm-0' })}>September reading</h2>
      <p className="m-0">24px inside the surface. 16px between these items.</p>
    </Surface>
  )
}
