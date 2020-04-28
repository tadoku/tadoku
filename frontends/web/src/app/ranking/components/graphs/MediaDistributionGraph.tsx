import React, { useState } from 'react'
import {
  makeWidthFlexible,
  DiscreteColorLegend,
  RadialChart,
  Hint,
  RadialChartPoint,
} from 'react-vis'
import styled from 'styled-components'

import HintContainer from './HintContainer'
import { ContestLog } from '../../interfaces'
import { aggregateMediaDistribution } from '../../transform/graph'
import { formatScore } from '../../transform/format'

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
              <strong>{formatScore(selected.amount)}</strong> points from{' '}
              <strong>{selected.medium.toLowerCase()}</strong> (
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
