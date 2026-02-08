import { chartDatasetDefaults } from 'ui'
import { Chart as ChartJS, registerables } from 'chart.js'
import { Doughnut } from 'react-chartjs-2'

ChartJS.register(...registerables)

export default function DoughnutExample() {
  return (
    <Doughnut
      data={{
        labels: ['Book', 'Comic', 'Sentence', 'News'],
        datasets: [
          {
            ...chartDatasetDefaults,
            label: 'Score',
            data: [200, 300, 400, 500],
            weight: 20,
          },
        ],
      }}
    />
  )
}
