import { chartColors } from 'ui'
import {
  Chart as ChartJS,
  Tooltip,
  Legend,
  LinearScale,
  CategoryScale,
  BarElement,
  PointElement,
  LineElement,
  LineController,
  BarController,
  TimeScale,
} from 'chart.js'
import { Chart } from 'react-chartjs-2'
import 'chartjs-adapter-luxon'
import { faker } from '@faker-js/faker'

ChartJS.register(
  Tooltip,
  Legend,
  LinearScale,
  CategoryScale,
  BarElement,
  PointElement,
  LineElement,
  LineController,
  BarController,
  TimeScale,
)

export function ReadingActivityChart() {
  const labels = Array.from(Array(14).keys()).map(
    day => '2022-12-' + (day + 1).toString().padStart(2, '0'),
  )

  const comicData = labels.map(() =>
    faker.datatype.number({ min: 0, max: 1000 }),
  )
  const bookData = labels.map(() =>
    faker.datatype.number({ min: 0, max: 1000 }),
  )
  const cumulativeScore = comicData
    .map((comic, i) => comic + bookData[i])
    .reduce((acc, val) => {
      if (acc.length > 0) {
        val += acc[acc.length - 1]
      }
      acc.push(val)
      return acc
    }, [] as number[])

  return (
    <Chart
      type="bar"
      data={{
        labels,
        datasets: [
          {
            type: 'line' as const,
            label: 'Cumulative Score',
            borderColor: chartColors[chartColors.length - 1],
            borderWidth: 2,
            fill: false,
            data: cumulativeScore,
            yAxisID: 'yCumulative',
          },
          {
            type: 'bar' as const,
            label: 'Comic',
            backgroundColor: chartColors[0],
            data: comicData,
            yAxisID: 'yScore',
          },
          {
            type: 'bar' as const,
            label: 'Book',
            backgroundColor: chartColors[1],
            data: bookData,
            yAxisID: 'yScore',
          },
        ],
      }}
      options={{
        maintainAspectRatio: false,
        scales: {
          x: {
            stacked: true,
            type: 'time',
            time: {
              tooltipFormat: 'MMMM d',
              unit: 'day',
              displayFormats: {
                day: 'MMM dd',
              },
            },
          },
          yScore: {
            title: {
              text: 'Score',
              display: true,
            },
            stacked: true,
            position: 'left',
          },
          yCumulative: {
            title: {
              text: 'Cumulative score',
              display: true,
            },
            position: 'right',
            grid: {
              drawOnChartArea: false,
            },
          },
        },
      }}
    />
  )
}
