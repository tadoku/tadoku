import React, { useState } from 'react'
import { ContestLog } from '../../interfaces'
import { aggregateMediaDistribution } from '../../transform/graph'
import {
  makeWidthFlexible,
  DiscreteColorLegend,
  RadialChart,
  Hint,
  RadialChartPoint,
} from 'react-vis'
import styled from 'styled-components'
import Constants from '../../../ui/Constants'

interface Props {
  logs: ContestLog[]
}

const MediaDistributionGraph = ({ logs }: Props) => {
  const data = aggregateMediaDistribution(logs)
  const [selected, setSelected] = useState(
    undefined as undefined | RadialChartPoint,
  )

  return (
    <Container>
      <FlexibleRadialChart
        innerRadius={60}
        radius={90}
        getAngle={(d: { amount: number }) => d.amount}
        getLabel={(d: { medium: string }) => d.medium}
        data={data.aggregated}
        width={200}
        height={200}
        padAngle={0.04}
        onValueMouseOver={(v: RadialChartPoint) => setSelected(v)}
        onSeriesMouseOut={() => setSelected(undefined)}
        colorType="literal"
      >
        {selected && (
          <Hint value={selected}>
            <HintContainer>
              {selected.amount} points from {selected.medium.toLowerCase()} (
              {Math.floor((selected.amount / data.totalAmount) * 100)}%)
            </HintContainer>
          </Hint>
        )}
      </FlexibleRadialChart>
      <DiscreteColorLegend
        items={data.legend}
        orientation="horizontal"
        height={60}
        style={{ textAlign: 'center' }}
      />
    </Container>
  )
}

export default MediaDistributionGraph

const FlexibleRadialChart = makeWidthFlexible(RadialChart)

const Container = styled.div`
  display: flex;
  flex-direction: column;
  align-items: stretch;
  width: 200px;
`

const HintContainer = styled.div`
  background: ${Constants.colors.darkWithAlpha(0.9)};
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  color: ${Constants.colors.light};
  padding: 8px 12px;
  border-radius: 0;
`
