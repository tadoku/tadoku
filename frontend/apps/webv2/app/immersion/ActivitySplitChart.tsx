import { chartDatasetDefaults } from 'ui'
import {
  Chart as ChartJS,
  BarController,
  BarElement,
  LinearScale,
  Tooltip,
  Legend,
  CategoryScale,
} from 'chart.js'
import { Bar } from 'react-chartjs-2'
import { ActivitySplitScore } from '@app/immersion/api'

ChartJS.register(
  BarElement,
  BarController,
  Tooltip,
  Legend,
  LinearScale,
  CategoryScale,
)

interface Props {
  activities: ActivitySplitScore[]
}

export function ActivitySplitChart({ activities }: Props) {
  return (
    <Bar
      options={{
        plugins: {
          legend: {
            display: false,
          },
          tooltip: {
            cornerRadius: 0,
          },
        },
        indexAxis: 'y' as const,
      }}
      data={{
        labels: activities.map(it => it.activity_name),
        datasets: [
          {
            barThickness: 30,
            ...chartDatasetDefaults,
            label: 'Score',
            data: activities.map(it => it.score),
          },
        ],
      }}
    />
  )
}
