import React from 'react'
import { ContestLog } from '../interfaces'
import styled from 'styled-components'
import { languageNameByCode, mediumDescriptionById } from '../database'
import { amountToPages } from '../transform/graph'
import { Button, ButtonContainer } from '../../ui/components'
import media from 'styled-media-query'
import Constants from '../../ui/Constants'
import { format } from 'date-fns'

interface Props {
  logs: ContestLog[]
  canEdit: boolean
  editLog: (log: ContestLog) => void
  deleteLog: (log: ContestLog) => void
}

const ContestLogsTable = (props: Props) => (
  <TableList>
    <Heading>
      <HeadingRow>
        <Column>Date</Column>
        <Column>Language</Column>
        <Column>Medium</Column>
        <Column>Description</Column>
        <Column alignRight>Amount</Column>
        <Column alignRight>Score</Column>
        {props.canEdit && <Column />}
      </HeadingRow>
    </Heading>
    <Body>
      {props.logs.map(l => (
        <Row key={l.id}>
          <Column title={l.date.toLocaleString()}>
            {format(l.date, 'MMMM do')}
          </Column>
          <Column>{languageNameByCode(l.languageCode)}</Column>
          <Column>{mediumDescriptionById(l.mediumId)}</Column>
          <Column>{l.description || 'N/A'}</Column>
          <Column alignRight>
            <strong>{amountToPages(l.amount)}</strong>
          </Column>
          <Column alignRight>
            <strong>{amountToPages(l.adjustedAmount)}</strong>
          </Column>
          {props.canEdit && (
            <Column style={{ width: '1px', whiteSpace: 'nowrap' }}>
              <ButtonContainer>
                <Button onClick={() => props.editLog(l)} icon="edit">
                  Edit
                </Button>
                <Button
                  onClick={() => props.deleteLog(l)}
                  icon="trash"
                  destructive
                >
                  Delete
                </Button>
              </ButtonContainer>
            </Column>
          )}
        </Row>
      ))}
    </Body>
  </TableList>
)

export default ContestLogsTable

const TableList = styled.table`
  padding: 0;
  width: 100%;
  background: ${Constants.colors.light};
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  border-collapse: collapse;

  ${media.lessThan('medium')`
    display: none;
  `}
`

const Heading = styled.thead`
  height: 55px;
  font-size: 16px;
  font-weight: bold;
  text-transform: uppercase;
  color: ${Constants.colors.nonFocusText};
`

const HeadingRow = styled.tr`
  margin: 20px 0;
  height: 55px;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.2)};
`

const Row = styled.tr`
  margin: 20px 0;
  padding: 20px 30px;

  &:nth-child(2n + 1) {
    background-color: ${Constants.colors.nonFocusTextWithAlpha(0.05)};
  }
`

const Column = styled.td`
  padding: 10px 20px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.04);
  text-align: ${({ alignRight }: { alignRight?: boolean }) =>
    alignRight ? 'right' : 'left'};
`

const Body = styled.tbody``
