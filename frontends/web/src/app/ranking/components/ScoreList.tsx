import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'

import { SubHeading } from '../../ui/components'
import { RankingRegistrationOverview } from '../interfaces'
import { scoreLabel, formatScore } from '../transform/format'

interface Props {
  registrationOverview: RankingRegistrationOverview
}

const ScoreList = ({ registrationOverview }: Props) => {
  return (
    <Scores>
      {registrationOverview.registrations.map(r => (
        <Score key={r.languageCode}>
          <ScoreLabel>{scoreLabel(r.languageCode)}</ScoreLabel>
          <ScoreValue>{formatScore(r.amount)}</ScoreValue>
        </Score>
      ))}
    </Scores>
  )
}

export default ScoreList

const Scores = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  margin-bottom: 20px;
  width: 100%;
  flex-wrap: wrap;

  ${media.lessThan('small')`
    flex-direction: column;
  `}
`

const Score = styled.div`
  width: 20%;
  height: 100px;

  ${media.lessThan('medium')`
    width: 50%;
  `}
`
const ScoreLabel = styled(SubHeading)`
  margin: 0;
`
const ScoreValue = styled.div`
  font-size: 38px;
  font-weight: bold;
  margin-top: 10px;
`
