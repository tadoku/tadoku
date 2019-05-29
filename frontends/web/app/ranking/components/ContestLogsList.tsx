import React from 'react'
import { ContestLog } from '../interfaces'
import styled from 'styled-components'
import { languageNameByCode, mediumDescriptionById } from '../database'
import { amountToPages } from '../transform'

const List = styled.table`
  list-style: none;
  padding: 0;
  margin: 0 auto;
  width: 100%;
  border-collapse: collapse;
`

const Row = styled.tr`
  margin: 20px 0;
  padding: 20px 30px;
  background-color: ${({ even }: { even?: boolean }) =>
    even ? 'rgba(0, 0, 0, 0.02)' : 'transparant'};
`

const Column = styled.td`
  padding: 15px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.04);
  text-align: ${({ alignRight }: { alignRight?: boolean }) =>
    alignRight ? 'right' : 'left'};
`

const Heading = styled.thead`
  font-weight: bold;
  font-size: 1.2em;

  td {
    border-bottom: 1px solid rgba(0, 0, 0, 0.1);
  }
`

const Body = styled.tbody``

interface Props {
  logs: ContestLog[]
}

const ContestLogList = (props: Props) => (
  <>
    <h1>Updates</h1>
    <List>
      <Heading>
        <Row>
          <Column>Date</Column>
          <Column>Language</Column>
          <Column>Medium</Column>
          <Column alignRight>Amount</Column>
          <Column alignRight>Score</Column>
        </Row>
      </Heading>
      <Body>
        {props.logs.map((l, i) => (
          <Row even={i % 2 === 0}>
            <Column>{l.date.toLocaleString()}</Column>
            <Column>{languageNameByCode(l.languageCode)}</Column>
            <Column>{mediumDescriptionById(l.mediumId)}</Column>
            <Column alignRight>{amountToPages(l.amount)}</Column>
            <Column alignRight>{amountToPages(l.adjustedAmount)}</Column>
          </Row>
        ))}
      </Body>
    </List>
  </>
)

export default ContestLogList
