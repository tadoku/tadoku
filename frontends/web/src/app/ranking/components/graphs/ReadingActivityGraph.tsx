import React, { useState } from 'react'
import { ContestLog } from '../../interfaces'
import { Contest } from '../../../contest/interfaces'
import {
  aggregateReadingActivity,
  prettyDate,
  amountToString,
} from '../../transform/graph'
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
import styled from 'styled-components'
import Constants from '../../../ui/Constants'
import { graphColor } from '../../../ui/components/Graphs'
import { format } from 'date-fns'

interface Props {
  logs: ContestLog[]
  contest: Contest
}

const ReadingActivityGraph = ({ logs, contest }: Props) => {
  const data = aggregateReadingActivity(logs, contest)
  const [selected, setSelected] = useState(
    undefined as undefined | LineMarkSeriesPoint,
  )

  return (
    <Container>
      <FlexiblePlot
        height={220}
        xType="ordinal"
        stackBy="y"
        margin={{ top: 0, bottom: 50, right: 0, left: 40 }}
      >
        <HorizontalGridLines />
        <VerticalGridLines />
        <XAxis
          tickFormat={raw => {
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
              {amountToString(selected.y as number)} in {selected.language} on
              <br />
              {prettyDate(new Date(selected.x))}
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

const HintContainer = styled.div`
  background: ${Constants.colors.darkWithAlpha(0.9)};
  box-shadow: 0px 2px 7px 1px rgba(0, 0, 0, 0.25);
  color: ${Constants.colors.light};
  padding: 8px 12px;
  border-radius: 4px;
`
