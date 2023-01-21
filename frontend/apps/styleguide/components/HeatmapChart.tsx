import { Chart as ChartJS, registerables } from 'chart.js'
import { Chart } from 'react-chartjs-2'
import { MatrixElement, MatrixController } from 'chartjs-chart-matrix'
import { color } from 'chart.js/helpers'
import 'chartjs-adapter-luxon'
import { faker } from '@faker-js/faker'

ChartJS.register(MatrixElement, MatrixController)

function isoDayOfWeek(dt: Date) {
  let wd = dt.getDay() // 0..6, from sunday
  wd = ((wd + 6) % 7) + 1 // 1..7 from monday
  return '' + wd // string so it gets parsed
}

function startOfToday() {
  return
}

function generateData() {
  const data = []
  const end = new Date(2023, 1, 1, 0, 0, 0, 0)
  let dt = new Date(2022, 1, 1, 0, 0, 0, 0)
  while (dt <= end) {
    const iso = dt.toISOString().substr(0, 10)
    data.push({
      x: iso,
      y: isoDayOfWeek(dt),
      d: iso,
      v: Math.random() * 50,
    })
    dt = new Date(dt.setDate(dt.getDate() + 1))
  }
  console.log(data)
  return data
}

const data = generateData()

function HeatmapChart() {
  return (
    <Chart
      type="matrix"
      data={{
        datasets: [
          {
            label: 'My Matrix',
            data,
            backgroundColor(c) {
              const value = c.dataset.data[c.dataIndex].v
              const alpha = (10 + value) / 60
              return color('green').alpha(alpha).rgbString()
            },
            borderColor(c) {
              const value = c.dataset.data[c.dataIndex].v
              const alpha = (10 + value) / 60
              return color('green').alpha(alpha).darken(0.3).rgbString()
            },
            borderWidth: 1,
            hoverBackgroundColor: 'yellow',
            hoverBorderColor: 'yellowgreen',
            width(c) {
              const a = c.chart.chartArea || {}
              return (a.right - a.left) / 53 - 1
            },
            height(c) {
              const a = c.chart.chartArea || {}
              return (a.bottom - a.top) / 7 - 1
            },
          },
        ],
      }}
      options={{
        aspectRatio: 5,
        plugins: {
          legend: {
            display: false,
          },
          tooltip: {
            displayColors: false,
            callbacks: {
              title() {
                return ''
              },
              label(context) {
                const v = context.dataset.data[context.dataIndex]
                return ['d: ' + v.d, 'v: ' + v.v.toFixed(2)]
              },
            },
          },
        },
        scales: {
          y: {
            type: 'time',
            offset: true,
            time: {
              unit: 'day',
              round: 'day',
              parser: 'd',
              displayFormats: {
                day: 'ccc',
              },
            },
            reverse: true,
            position: 'right',
            ticks: {
              maxRotation: 0,
              autoSkip: false,
              padding: 1,
              font: {
                size: 9,
              },
            },
            grid: {
              display: false,
              tickLength: 0,
            },
          },
          x: {
            type: 'time',
            position: 'bottom',
            offset: true,
            time: {
              unit: 'week',
              round: 'week',

              displayFormats: {
                month: 'MMMM',
              },
            },
            ticks: {
              maxRotation: 0,
              autoSkip: false,
              font: {
                size: 9,
              },
            },
            grid: {
              display: false,
              tickLength: 0,
            },
          },
        },
      }}
    />
  )
}

export default HeatmapChart
