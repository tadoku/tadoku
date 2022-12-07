import { datasetDefaults } from '@components/charts'
import { CodeBlock, Preview, Separator, Title } from '@components/example'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'react-chartjs-2'

ChartJS.register(ArcElement, Tooltip, Legend)

export default function Typography() {
  return (
    <>
      <h1 className="title mb-8">Charts</h1>
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
