import { chartColors, datasetDefaults } from '@components/charts'
import { CodeBlock, Preview, Separator, Title } from '@components/example'
import {
  Chart as ChartJS,
  ArcElement,
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
import { Doughnut } from 'react-chartjs-2'
import 'chartjs-adapter-luxon'
import { faker } from '@faker-js/faker'

ChartJS.register(
  ArcElement,
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

export default function Typography() {
  return (
    <>
      <h1 className="title mb-8">Charts</h1>

      <Title>Reading activity chart</Title>
      <Preview>
        <div className="max-w-[900px]">
          <ReadingActivityChart />
        </div>
      </Preview>
      <CodeBlock
        language="html"
        code={`import { chartColors, datasetDefaults } from '@components/charts'
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

function ReadingActivityChart() {
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
}`}
      />

      <Separator />

      <Title>Doughnut chart</Title>

      <div className="h-stack w-full">
        <div className="w-96">
          <Preview>
            <Doughnut
              data={{
                labels: ['Book', 'Comic', 'Sentence', 'News'],
                datasets: [
                  {
                    ...datasetDefaults,
                    label: 'Score',
                    data: [200, 300, 400, 500],
                    weight: 20,
                  },
                ],
              }}
            />
          </Preview>
        </div>
        <div className="flex-1">
          <CodeBlock
            code={`import { datasetDefaults } from '@components/charts'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'react-chartjs-2'

ChartJS.register(ArcElement, Tooltip, Legend)

const Example = () => (
  <Doughnut
    data={{
      labels: ['Book', 'Comic', 'Sentence', 'News'],
      datasets: [
        {
          ...datasetDefaults,
          label: 'Score',
          data: [200, 300, 400, 500],
          weight: 20,
        },
      ],
    }}
  />
)`}
          />
        </div>
      </div>
    </>
  )
}

function ReadingActivityChart() {
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
