import React from 'react'
import { format } from 'date-fns'
import styled from 'styled-components'
import media from 'styled-media-query'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

import { Contest } from '@app/contest/interfaces'
import Constants from '@app/ui/Constants'

interface Props {
  contest: Contest
}

const ContestPeriod = ({ contest }: Props) => {
  return (
    <Container>
      <Date>
        <Label>Starting</Label>
        <Value>{format(contest.start, 'MMMM dd')}</Value>
      </Date>
      <Icon icon="arrow-right" />
      <Date>
        <Label>Ending</Label>
        <Value>{format(contest.end, 'MMMM dd')}</Value>
      </Date>
    </Container>
  )
}

export default ContestPeriod

const Container = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  box-sizing: border-box;
  background: ${Constants.colors.light};

  ${media.lessThan('small')`width: 100%; margin-top: 5px;`}
`

const Icon = styled(FontAwesomeIcon)`
  color: ${Constants.colors.nonFocusText};
  opacity: 0.4;
  margin: 0 20px;
`

const Date = styled.div``

const Label = styled.div`
  font-size: 12px;
  text-transform: uppercase;
  font-weight: bold;
  color: ${Constants.colors.nonFocusText};
`
const Value = styled.div`
  font-weight: bold;
  font-size: 13px;
`
