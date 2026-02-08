import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import HeatmapExample from '@examples/charts/heatmap'
import heatmapCode from '@examples/charts/heatmap.tsx?raw'

import ReadingActivityChart from '@examples/charts/reading-activity'
import readingActivityCode from '@examples/charts/reading-activity.tsx?raw'

import DoughnutExample from '@examples/charts/doughnut'
import doughnutCode from '@examples/charts/doughnut.tsx?raw'

export default function Charts() {
  return (
    <>
      <h1 className="title mb-8">Charts</h1>

      <Showcase title="Heatmap" code={heatmapCode}>
        <div className="max-w-[900px]">
          <HeatmapExample />
        </div>
      </Showcase>

      <Separator />

      <Showcase title="Reading activity chart" code={readingActivityCode}>
        <div className="max-w-[900px]">
          <ReadingActivityChart />
        </div>
      </Showcase>

      <Separator />

      <Showcase title="Doughnut chart" code={doughnutCode}>
        <div className="w-96">
          <DoughnutExample />
        </div>
      </Showcase>
    </>
  )
}
