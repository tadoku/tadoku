import React, { useEffect, useState } from 'react'
import { format } from 'date-fns'
import styled from 'styled-components'
import media from 'styled-media-query'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { formatDistanceToNow } from 'date-fns'

import { Contest } from '@app/contest/interfaces'
import Constants from '@app/ui/Constants'

interface Props {
  contest: Contest
}

const ContestPeriod = ({ contest }: Props) => {
  const now = new Date()
  const startingLabel = now > contest.start ? 'Started' : 'Starting'
  const endingLabel = now > contest.end ? 'Ended' : 'Ending'

  return (
    <Container>
      <Dates>
        <Box>
          <Label>{startingLabel}</Label>
          <Value>{format(contest.start, 'MMMM dd')}</Value>
        </Box>
        <Icon icon="arrow-right" />
        <Box>
          <Label>{endingLabel}</Label>
          <Value>{format(contest.end, 'MMMM dd')}</Value>
        </Box>
      </Dates>
      <RemainingTime contest={contest} />
    </Container>
  )
}

export default ContestPeriod

const RemainingTime = ({ contest }: Props) => {
  const [now, setNow] = useState(() => new Date())

  useEffect(() => {
    const id = setInterval(() => {
      setNow(new Date())
    }, 1000)
    return () => clearInterval(id)
  }, [])

  const startingLabel = formatDistanceToNow(contest.start, {
    includeSeconds: true,
  })
  const endingLabel = formatDistanceToNow(contest.end, {
    includeSeconds: true,
  })

  if (now < contest.start) {
    return (
      <RemainingBlock>
        Starts in <strong>{startingLabel}</strong>
      </RemainingBlock>
    )
  }

  if (now < contest.end) {
    return (
      <RemainingBlock>
        Ends in <strong>{endingLabel}</strong>
      </RemainingBlock>
    )
  }

  return <RemainingBlock>Contest has ended</RemainingBlock>
}

const RemainingBlock = styled.div`
  font-size: 12px;
  margin: 10px -15px -10px -15px;
  padding: 5px 15px;
  border-top: 1px solid ${Constants.colors.lightGray};
`

const Container = styled.div`
  padding: 10px 15px;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  box-sizing: border-box;
  background: ${Constants.colors.light};
  width: 100%;

  ${media.lessThan('small')`margin-top: 5px;`}
`

const Dates = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
`

const Icon = styled(FontAwesomeIcon)`
  color: ${Constants.colors.nonFocusText};
  opacity: 0.4;
  margin: 0 20px;
`

const Box = styled.div``

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
