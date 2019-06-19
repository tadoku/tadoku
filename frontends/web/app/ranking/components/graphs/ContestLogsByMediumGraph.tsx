import React from 'react'
import { ContestLog } from '../../interfaces'
import { aggregateContestLogsByMedium } from '../../transform'
import { makeWidthFlexible, DiscreteColorLegend, RadialChart } from 'react-vis'
import styled from 'styled-components'
import Constants from '../../../ui/Constants'

interface Props {
  logs: ContestLog[]
}

const Graph = ({ logs }: Props) => {
  const data = aggregateContestLogsByMedium(logs)

  return (
    <Container>
      <FlexibleRadialChart
        innerRadius={100}
        radius={140}
        getAngle={(d: { theta: number }) => d.theta}
        data={data.aggregated}
        width={300}
        height={300}
        padAngle={0.04}
      />
      <DiscreteColorLegend
        items={data.legend}
        orientation="horizontal"
        height={60}
      />
    </Container>
  )
}

export default Graph

const FlexibleRadialChart = makeWidthFlexible(RadialChart)

const Container = styled.div`
  display: flex;
  flex-direction: column;
  align-items: stretch;
`
