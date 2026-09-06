import { Surface } from 'paper-ui'

export default function Summary() {
  return (
    <Surface as="article" elevation="flat" className="paper-stack">
      <h2 className="paper-type-component">September reading</h2>
      <p>128 pages across three sessions.</p>
    </Surface>
  )
}
