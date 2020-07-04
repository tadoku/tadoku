import React, { useState, useMemo } from 'react'
import {
  XYPlot,
  XAxis,
  YAxis,
  HorizontalGridLines,
  VerticalGridLines,
  DiscreteColorLegend,
  VerticalBarSeries,
  Hint,
  LineMarkSeriesPoint,
  makeWidthFlexible,
  ChartLabel,
} from 'react-vis'
import { format } from 'date-fns'
import styled from 'styled-components'

import { ContestLog } from '../../interfaces'
import { Contest } from '@app/contest/interfaces'
import { aggregateReadingActivity } from '../../transform/graph'
import { graphColor } from '@app/ui/components/Graphs'
import HintContainer from './HintContainer'
import { formatLineMarkSeriesPoint } from '../../transform/format'

interface Props {
  logs: ContestLog[]
  contest: Contest
  effectCount: number
}

const ReadingActivityGraph = ({ logs, contest, effectCount }: Props) => {
  const data = useMemo(() => aggregateReadingActivity(logs, contest), [
    effectCount,
    logs,
    contest,
  ])
  const [selected, setSelected] = useState(
    undefined as undefined | LineMarkSeriesPoint,
  )

  return (
    <Container>
      <FlexiblePlot
        height={220}
        xType="ordinal"
        stackBy="y"
        margin={{ top: 5, bottom: 50, right: 0, left: 45 }}
      >
        <HorizontalGridLines />
        <VerticalGridLines />
        <XAxis
          tickFormat={(raw, i) => {
            const length =
              data.aggregated[Object.keys(data.aggregated)[0]].length
            if (length > 40 && i % 3 != 0) {
              return ''
            }
            if (length > 20 && length < 40 && i % 2 != 0) {
              return ''
            }
            return `${format(new Date(raw), 'MMM dd')}`
          }}
          tickLabelAngle={-55}
        />
        <YAxis />
        {Object.keys(data.aggregated).map((language, i) => (
          <VerticalBarSeries
            data={data.aggregated[language] as any[]}
            key={language}
            color={graphColor(i)}
            onValueMouseOver={value => setSelected(value)}
            onValueMouseOut={() => setSelected(undefined)}
          />
        ))}
        {selected && (
          <Hint value={selected}>
            <HintContainer>
              <strong>{formatLineMarkSeriesPoint(selected)}</strong> in{' '}
              <strong>{selected.language}</strong> on
              <br />
              {format(new Date(selected.x), 'MMMM do')}
            </HintContainer>
          </Hint>
        )}
        <ChartLabel
          text="Date"
          includeMargin={false}
          xPercent={0.95}
          yPercent={0.95}
          style={{
            stroke: 'white',
            opacity: 1,
            strokeWidth: '3',
            fontWeight: 'bold',
          }}
        />
        <ChartLabel
          text="Date"
          includeMargin={false}
          xPercent={0.95}
          yPercent={0.95}
          style={{
            fontWeight: 'bold',
          }}
        />
        <ChartLabel
          text="Score"
          includeMargin={false}
          xPercent={0.01}
          yPercent={0.1}
          style={{
            stroke: 'white',
            opacity: 1,
            strokeWidth: '3',
            fontWeight: 'bold',
          }}
        />
        <ChartLabel
          text="Score"
          includeMargin={false}
          xPercent={0.01}
          yPercent={0.1}
          style={{
            fontWeight: 'bold',
          }}
        />
      </FlexiblePlot>
      <DiscreteColorLegend
        items={data.legend}
        orientation="horizontal"
        height={52}
      />
    </Container>
  )
}

export default ReadingActivityGraph

const FlexiblePlot = makeWidthFlexible(XYPlot)

const Container = styled.div`
  display: flex;
  flex-direction: column;
  align-items: stretch;
`
