import React, { useState } from 'react'
import { ContestLog, RankingRegistrationOverview } from '../interfaces'
import styled from 'styled-components'
import { languageNameByCode, mediumDescriptionById } from '../database'
import { amountToPages } from '../transform'
import EditLogFormModal from './modals/EditLogFormModal'
import { State } from '../../store'
import { connect } from 'react-redux'
import { User } from '../../session/interfaces'
import { Button, ButtonContainer } from '../../ui/components'
import RankingApi from '../api'
import media from 'styled-media-query'
import ContestLogsTable from './ContestLogsTable'

interface Props {
  logs: ContestLog[]
  canEdit: boolean
  editLog: (log: ContestLog) => void
  deleteLog: (log: ContestLog) => void
}

const ContestLogsList = (props: Props) => (
  <List>
    {props.logs.map((l, i) => (
      <Item even={i % 2 === 0} key={l.id}>
        <span>{l.date.toLocaleString()}</span>
        <span>{languageNameByCode(l.languageCode)}</span>
        <span>{mediumDescriptionById(l.mediumId)}</span>
        <span>{amountToPages(l.amount)}</span>
        <span>{amountToPages(l.adjustedAmount)}</span>
        {props.canEdit && (
          <span style={{ width: '1px', whiteSpace: 'nowrap' }}>
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
          </span>
        )}
      </Item>
    ))}
  </List>
)

export default ContestLogsList

const List = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0 auto;
  width: 100%;

  ${media.greaterThan('medium')`
    display: none;
  `}
`

const Item = styled.li`
  margin: 20px 0;
  padding: 20px 30px;
  background-color: ${({ even }: { even?: boolean }) =>
    even ? 'rgba(0, 0, 0, 0.02)' : 'transparant'};
`
