import React from 'react'
import { ContestLog } from '../interfaces'
import styled from 'styled-components'
import { formatLanguageName, formatMediaDescription } from '../transform/format'
import { Button, ButtonContainer } from '../../ui/components'
import media from 'styled-media-query'
import Constants from '../../ui/Constants'
import { format } from 'date-fns'
import { formatScore } from '../transform/format'

interface Props {
  logs: ContestLog[]
  canEdit: boolean
  editLog: (log: ContestLog) => void
  deleteLog: (log: ContestLog) => void
}

const UpdatesList = (props: Props) => (
  <TableList>
    <Heading>
      <HeadingRow>
        <HideSmallColumn>Date</HideSmallColumn>
        <HideSmallColumn>Language</HideSmallColumn>
        <HideSmallColumn>Medium</HideSmallColumn>
        <HideSmallColumn>Description</HideSmallColumn>
        <HideSmallColumn alignRight>Amount</HideSmallColumn>
        <HideSmallColumn alignRight>Score</HideSmallColumn>
        <ShowSmallColumn>Description</ShowSmallColumn>
        {props.canEdit && <Column />}
      </HeadingRow>
    </Heading>
    <Body>
      {props.logs.map(l => (
        <Row key={l.id}>
          <HideSmallColumn
            style={{ whiteSpace: 'nowrap' }}
            title={l.date.toLocaleString()}
          >
            {format(l.date, 'MMM do')}
          </HideSmallColumn>
          <HideSmallColumn>
            {formatLanguageName(l.languageCode)}
          </HideSmallColumn>
          <HideSmallColumn>
            {formatMediaDescription(l.mediumId)}
          </HideSmallColumn>
          <DescriptionColumn title={l.description || 'N/A'}>
            <span>{l.description || 'N/A'}</span>
          </DescriptionColumn>
          <HideSmallColumn alignRight>
            <strong>{formatScore(l.amount)}</strong>
          </HideSmallColumn>
          <HideSmallColumn alignRight>
            <strong>{formatScore(l.adjustedAmount)}</strong>
          </HideSmallColumn>
          <ShowSmallColumn>
            <strong>{formatScore(l.amount)}</strong> of{' '}
            <strong>{formatMediaDescription(l.mediumId)}</strong> in{' '}
            <strong>{formatLanguageName(l.languageCode)}</strong> at{' '}
            <strong>{format(l.date, 'MMM do')}</strong> for a total of{' '}
            <strong>{formatScore(l.adjustedAmount)}</strong> points
          </ShowSmallColumn>
          {props.canEdit && (
            <Column style={{ width: '1px', whiteSpace: 'nowrap', padding: 0 }}>
              <ActionButtonContainer>
                <Button onClick={() => props.editLog(l)} icon="edit">
                  <span>Edit</span>
                </Button>
                <Button
                  onClick={() => props.deleteLog(l)}
                  icon="trash"
                  destructive
                >
                  <span>Delete</span>
                </Button>
              </ActionButtonContainer>
            </Column>
          )}
        </Row>
      ))}
    </Body>
  </TableList>
)

export default UpdatesList

const TableList = styled.table`
  padding: 0;
  width: 100%;
  background: ${Constants.colors.light};
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  border-collapse: collapse;
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

const Column = styled.td<{ alignRight?: boolean }>`
  padding: 10px 20px;
  text-align: ${({ alignRight }) => (alignRight ? 'right' : 'left')};
`

const HideSmallColumn = styled(Column)`
  ${media.lessThan('medium')`
    display: none;
  `}
`

const ShowSmallColumn = styled(Column)`
  display: none;

  ${media.lessThan('medium')`
    display: table-cell;
  `}
`

const DescriptionColumn = styled(HideSmallColumn)`
  position: relative;
  height: 69px;
  box-sizing: border-box;
  width: 100%;

  &:before {
    content: '&nbsp;';
    visibility: hidden;
  }

  span {
    position: absolute;
    height: 69px;
    line-height: 69px;
    top: 0;
    left: 20px;
    right: 20px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
`

const Body = styled.tbody``

const ActionButtonContainer = styled(ButtonContainer)`
  ${media.lessThan('medium')`
    button > span { display: none; }
    button svg { margin: 0; }
  `}
`
