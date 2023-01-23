import { chartDatasetDefaults } from 'ui'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'react-chartjs-2'
import { ActivitySplitScore } from '@app/immersion/api'

ChartJS.register(ArcElement, Tooltip, Legend)

interface Props {
  activities: ActivitySplitScore[]
}

export function ActivitySplitChart({ activities }: Props) {
  return (
    <Doughnut
      options={{
        plugins: {
          legend: {
            position: 'bottom',
            align: 'start',
          },
        },
      }}
      data={{
        labels: activities.map(it => it.activity_name),
        datasets: [
          {
            ...chartDatasetDefaults,
            label: 'Score',
            data: activities.map(it => it.score),
          },
        ],
      }}
    />
  )
}
