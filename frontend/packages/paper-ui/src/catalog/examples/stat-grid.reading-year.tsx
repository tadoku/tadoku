import { StatGrid } from 'paper-ui'

export default function Example() {
  return (
    <StatGrid
      stats={[
        { label: 'Personal score', value: '1,240.5' },
        { label: 'Entries', value: '18', detail: '12 active days' },
        { label: 'Languages', value: '2', detail: 'Japanese leads' },
        { label: 'Contests', value: '1' },
      ]}
    />
  )
}
