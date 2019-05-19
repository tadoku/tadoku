import React from 'react'
import { ContestLog } from '../interfaces'
import { aggregateContestLogsByDays } from '../transform'
import {
  XYPlot,
  XAxis,
  YAxis,
  HorizontalGridLines,
  VerticalGridLines,
  LineSeries,
} from 'react-vis'

interface Props {
  logs: ContestLog[]
}

const Graph = ({ logs }: Props) => {
  const data = aggregateContestLogsByDays(logs, {
    id: 1,
    start: new Date('2019-05-01'),
    end: new Date('2019-05-31'),
    open: true,
  })

  return (
    <XYPlot width={1000} height={400} xType={'time'}>
      <HorizontalGridLines />
      <VerticalGridLines />
      <XAxis title="Days" tickFormat={date => date.getDate()} />
      <YAxis title="Pages" />

      {Object.keys(data).map(language => (
        <LineSeries data={data[language] as any[]} key={language} />
      ))}
    </XYPlot>
  )
}

export default Graph
