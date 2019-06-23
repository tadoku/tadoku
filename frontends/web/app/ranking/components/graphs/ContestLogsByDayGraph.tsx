import React, { useState } from 'react'
import { ContestLog } from '../../interfaces'
import { Contest } from '../../../contest/interfaces'
import {
  aggregateContestLogsByDays,
  prettyDate,
  amountToString,
} from '../../transform'
import {
  XYPlot,
  XAxis,
  YAxis,
  AreaSeries,
  HorizontalGridLines,
  VerticalGridLines,
  makeWidthFlexible,
  DiscreteColorLegend,
  LineMarkSeries,
  Hint,
  LineMarkSeriesPoint,
} from 'react-vis'
import styled from 'styled-components'
import Constants from '../../../ui/Constants'
import {
  GradientDefinitions,
  graphColor,
  gradientDefinitionUrl,
} from '../../../ui/components/Graphs'

interface Props {
  logs: ContestLog[]
  contest: Contest
}

const Graph = ({ logs, contest }: Props) => {
  const data = aggregateContestLogsByDays(logs, contest)
  const [selected, setSelected] = useState(undefined as
    | undefined
    | LineMarkSeriesPoint)

  return (
    <Container>
      <FlexiblePlot height={400} xType={'time'}>
        {GradientDefinitions}
        <HorizontalGridLines />
        <VerticalGridLines />
        <XAxis
          title="Date"
          tickFormat={date => `${date.getMonth() + 1}-${date.getDate()}`}
        />
        <YAxis title="Pages" />
        {Object.keys(data.aggregated).map((language, i) => (
          <AreaSeries
            data={data.aggregated[language] as any[]}
            key={language}
            curve={'curveMonotoneX'}
            onValueMouseOver={value => setSelected(value)}
            onValueMouseOut={() => setSelected(undefined)}
            color={gradientDefinitionUrl(i)}
            opacity={0.3}
          />
        ))}
        {Object.keys(data.aggregated).map((language, i) => (
          <LineMarkSeries
            data={data.aggregated[language] as any[]}
            curve={'curveMonotoneX'}
            onValueMouseOver={value => setSelected(value)}
            onValueMouseOut={() => setSelected(undefined)}
            key={language}
            color={graphColor(i)}
            sizeRange={[1, 4]}
          />
        ))}
        {selected && (
          <Hint value={selected}>
            <HintContainer>
              {amountToString(selected.y as number)} in {selected.language} on
              <br />
              {prettyDate(selected.x as Date)}
            </HintContainer>
          </Hint>
        )}
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

const HintContainer = styled.div`
  background: ${Constants.colors.darkWithAlpha(0.9)};
  box-shadow: 0px 2px 7px 1px rgba(0, 0, 0, 0.25);
  color: ${Constants.colors.light};
  padding: 8px 12px;
  border-radius: 4px;
`
