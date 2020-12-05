import React from 'react'
import styled from 'styled-components'
import { formatDistanceToNow } from 'date-fns'

import { ContestLog } from '../interfaces'
import Constants from '@app/ui/Constants'
import {
  formatLanguageName,
  formatMediaDescription,
  formatScore,
} from '../transform/format'
import media from 'styled-media-query'

interface Props {
  logs: ContestLog[]
  loading: boolean
}

const ContestUpdates = ({ logs, loading }: Props) => {
  if (loading || logs.length === 0) {
    return null
  }

  const grouped = logs.reduce((accumulator, log) => {
    const { length } = accumulator
    if (length === 0) {
      accumulator.push([log])
      return accumulator
    }
    const previous = accumulator[length - 1]
    if (previous[0].userId === log.userId) {
      previous.push(log)
      return accumulator
    } else {
      accumulator.push([log])
      return accumulator
    }
  }, [] as ContestLog[][])

  return (
    <Container>
      <h3>Recent updates</h3>
      <List>
        {grouped.map(logs => (
          <Update logs={logs} key={logs[0].id} />
        ))}
      </List>
    </Container>
  )
}

export default ContestUpdates

const List = styled.ul`
  padding: 0;
  width: 100%;
  margin: 0;
  list-style: none;
`

const Update = ({ logs }: { logs: ContestLog[] }) => {
  if (logs.length === 0) {
    return null
  }

  const totalScore = logs
    .map(l => l.amount)
    .reduce((total, amount) => total + amount)

  return (
    <StyledUpdate>
      <Header>
        <DisplayName>{logs[0].userDisplayName}</DisplayName>
        <Score>+{formatScore(totalScore)}</Score>
      </Header>
      {logs.map(log => (
        <Details key={log.id}>
          <Description>
            {formatScore(log.amount)} of {formatMediaDescription(log.mediumId)}{' '}
            in {formatLanguageName(log.languageCode)}
          </Description>
          <When>
            {formatDistanceToNow(log.date, {
              includeSeconds: true,
            })}{' '}
            ago
          </When>
        </Details>
      ))}
    </StyledUpdate>
  )
}

const Container = styled.div`
  ${media.lessThan('medium')`
    display: none;
  `}
`

const StyledUpdate = styled.li`
  background: ${Constants.colors.light};
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  border-collapse: collapse;

  & + & {
    margin-top: 10px;
  }
`

const Header = styled.div`
  display: flex;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.07)};
  padding: 10px 15px;
`
const Details = styled.div`
  background-color: ${Constants.colors.nonFocusTextWithAlpha(0.05)};
  padding: 10px 15px;

  & + & {
    border-top: 1px solid ${Constants.colors.nonFocusTextWithAlpha(0.07)};
  }
`
const DisplayName = styled.div`
  font-weight: bold;
  font-size: 1em;
  color: ${Constants.colors.dark};
  line-height: 1.4em;
`
const When = styled.div`
  color: ${Constants.colors.nonFocusText};
  font-size: 0.7em;
`
const Description = styled.div`
  font-size: 0.9em;
  line-height: 1.7em;
`
const Score = styled.div`
  text-align: right;
  flex: 1;
  font-weight: bold;
  color: ${Constants.colors.success};
`
