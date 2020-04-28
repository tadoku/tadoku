import React from 'react'
import { ContestLog } from '../interfaces'
import styled from 'styled-components'
import { formatLanguageName, formatMediaDescription } from '../transform/format'
import { Button, ButtonContainer } from '../../ui/components'
import media from 'styled-media-query'
import { formatScore } from '../transform/format'

interface Props {
  logs: ContestLog[]
  canEdit: boolean
  editLog: (log: ContestLog) => void
  deleteLog: (log: ContestLog) => void
}

const ContestLogsList = (props: Props) => (
  <List>
    {props.logs.map(log => (
      <Item key={log.id}>
        <UpdateText>
          <strong>{formatScore(log.amount)}</strong> of{' '}
          <strong>{formatMediaDescription(log.mediumId)}</strong> in{' '}
          <strong>{formatLanguageName(log.languageCode)}</strong> at{' '}
          <strong>{log.date.toLocaleString()}</strong> for a total of{' '}
          <strong>{formatScore(log.adjustedAmount)}</strong> points
        </UpdateText>
        {props.canEdit && (
          <ActionButtonContainer>
            <Button onClick={() => props.editLog(log)} icon="edit">
              Edit
            </Button>
            <Button
              onClick={() => props.deleteLog(log)}
              icon="trash"
              destructive
            >
              Delete
            </Button>
          </ActionButtonContainer>
        )}
      </Item>
    ))}
  </List>
)

export default ContestLogsList

const List = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0 -30px;

  ${media.greaterThan('medium')`
    display: none;
  `}
`

const Item = styled.li`
  margin: 10px 0;
  padding: 15px;
  border-radius: 2px;
  background-color: rgba(0, 0, 0, 0.03);
`

const UpdateText = styled.p`
  padding: 0;
  margin: 0 5px 15px;
`

const ActionButtonContainer = styled(ButtonContainer)`
  > button {
    flex: 1;
  }
`
