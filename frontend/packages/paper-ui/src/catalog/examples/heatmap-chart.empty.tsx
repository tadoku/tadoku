import { HeatmapChart } from 'paper-ui'

export default function Example() {
  return (
    <section aria-label="2023 reading activity">
      <p>No reading activity recorded in 2023.</p>
      <HeatmapChart id="empty-reading-activity" year={2023} data={[]} />
    </section>
  )
}
