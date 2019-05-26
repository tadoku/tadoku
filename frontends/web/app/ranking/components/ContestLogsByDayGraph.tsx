import React from 'react'
import { ContestLog } from '../interfaces'
import { Contest } from '../../contest/interfaces'
import { aggregateContestLogsByDays } from '../transform'
import {
  XYPlot,
  XAxis,
  YAxis,
  HorizontalGridLines,
  VerticalGridLines,
  LineSeries,
  makeWidthFlexible,
} from 'react-vis'

interface Props {
  logs: ContestLog[]
  contest: Contest
}

const FlexiblePlot = makeWidthFlexible(XYPlot)

const Graph = ({ logs, contest }: Props) => {
  const data = aggregateContestLogsByDays(logs, contest)

  return (
    <FlexiblePlot height={400} xType={'time'}>
      <HorizontalGridLines />
      <VerticalGridLines />
      <XAxis
        title="Days"
        tickFormat={date => `${date.getMonth() + 1}-${date.getDate()}`}
      />
      <YAxis title="Pages" />

      {Object.keys(data).map(language => (
        <LineSeries data={data[language] as any[]} key={language} />
      ))}
    </FlexiblePlot>
  )
}

export default Graph
