import { stack, Surface, typography } from 'paper-ui'

export default function Summary() {
  return (
    <Surface as="article" elevation="flat" className={stack()}>
      <h2 className={typography({ role: 'component' })}>September reading</h2>
      <p>128 pages across three sessions.</p>
    </Surface>
  )
}
