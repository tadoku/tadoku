import React from 'react'
import { ContestLog } from '../interfaces'
import styled from 'styled-components'
import { languageNameByCode, mediumDescriptionById } from '../database'
import { amountToPages } from '../transform'

const List = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0 auto;
`

const Row = styled.li`
  margin: 20px 0;
  padding: 20px 30px;
  border-radius: 2px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
`

interface Props {
  logs: ContestLog[]
}

const ContestLogList = (props: Props) => (
  <>
    <h1>Updates</h1>
    <List>
      {props.logs.map(l => (
        <Row>
          {languageNameByCode(l.languageCode)}: {amountToPages(l.amount)} pages
          of {mediumDescriptionById(l.mediumId)} (
          {amountToPages(l.adjustedAmount)})
        </Row>
      ))}
    </List>
  </>
)

export default ContestLogList
