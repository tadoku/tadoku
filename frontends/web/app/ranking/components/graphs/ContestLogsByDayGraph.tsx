import React from 'react'
import { ContestLog } from '../../interfaces'
import { Contest } from '../../../contest/interfaces'
import { aggregateContestLogsByDays } from '../../transform'
import {
  XYPlot,
  XAxis,
  YAxis,
  HorizontalGridLines,
  VerticalGridLines,
  makeWidthFlexible,
  DiscreteColorLegend,
  LineMarkSeries,
} from 'react-vis'
import styled from 'styled-components'

interface Props {
  logs: ContestLog[]
  contest: Contest
}

const Graph = ({ logs, contest }: Props) => {
  const data = aggregateContestLogsByDays(logs, contest)

  return (
    <Container>
      <FlexiblePlot height={400} xType={'time'}>
        <HorizontalGridLines />
        <VerticalGridLines />
        <XAxis
          title="Days"
          tickFormat={date => `${date.getMonth() + 1}-${date.getDate()}`}
        />
        <YAxis title="Pages" />
        {Object.keys(data.aggregated).map(language => (
          <LineMarkSeries
            data={data.aggregated[language] as any[]}
            curve={'curveMonotoneX'}
            key={language}
          />
        ))}
      </FlexiblePlot>
      <DiscreteColorLegend
        items={data.legend}
        orientation="horizontal"
        height={60}
      />
    </Container>
  )
}

export default Graph

const FlexiblePlot = makeWidthFlexible(XYPlot)

const Container = styled.div`
  display: flex;
  flex-direction: column;
  align-items: stretch;
`
