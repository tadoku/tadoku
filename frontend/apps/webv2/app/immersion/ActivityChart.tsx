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
import {
  ContestRegistrationView,
  useContestProfileReadingActivity,
} from '@app/immersion/api'
import { DateTime, Duration, Interval } from 'luxon'

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

interface Props {
  registration: ContestRegistrationView
  userId: string
}

export function ActivityChart({ userId, registration }: Props) {
  const activity = useContestProfileReadingActivity({
    userId,
    contestId: registration.contest_id,
  })

  if (activity.isLoading || activity.isIdle) {
    return <p>Loading...</p>
  }

  if (activity.isError || !registration.contest) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const end = DateTime.fromISO(registration.contest.contest_end)
    .plus(Duration.fromObject({ days: 1 }))
    .toISODate()

  const period = Interval.fromISO(
    `${registration.contest.contest_start}/${end}`,
  )
  const labels = period.splitBy({ day: 1 }).map(it => it.start.toISODate())
  const indexForLabel = labels.reduce((acc, label, i) => {
    acc[label] = i
    return acc
  }, {} as { [key: string]: number })

  const datasets = registration.languages.reduce(
    (acc, language, i) => {
      acc[language.code] = {
        code: language.code,
        label: language.name,
        data: labels.map(_ => 0),
        type: 'bar' as const,
        backgroundColor: chartColors[i],
        yAxisID: 'yScore',
      }
      return acc
    },
    {} as {
      [key: string]: {
        code: string
        label: string
        data: number[]
        type: 'bar'
        backgroundColor: string
        yAxisID: string
      }
    },
  )

  for (const row of activity.data.rows) {
    const i = indexForLabel[row.date]
    datasets[row.language_code].data[i] = row.score
  }

  const cumulativeScore = Object.values(datasets)
    .map(it => it.data)
    .reduce((acc, dataset) => {
      return acc.map((val, i) => val + dataset[i])
    }, Array(labels.length).fill(0) as number[])
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
          ...Object.values(datasets),
        ],
      }}
      options={{
        plugins: {
          tooltip: {
            cornerRadius: 0,
          },
        },
        maintainAspectRatio: false,
        scales: {
          x: {
            stacked: true,
            type: 'time',
            time: {
              tooltipFormat: 'MMMM d, yyyy',
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
