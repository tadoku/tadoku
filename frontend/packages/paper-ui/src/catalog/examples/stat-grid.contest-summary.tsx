import { StatGrid } from 'paper-ui'

export default function ContestSummaryExample() {
  return (
    <StatGrid
      variant="sheet"
      stats={[
        { label: 'Score', value: '20,969.2', detail: '4th place' },
        { label: 'Active days', value: '68', detail: '308.4 per active day' },
        { label: 'Logs', value: '127' },
        { label: 'Best day', value: '1,114.8', detail: '30 Sept' },
      ]}
    />
  )
}
