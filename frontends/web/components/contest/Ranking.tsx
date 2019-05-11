import React from 'react'
import { Ranking } from '../../domain/Ranking'

interface Props {
  rankings: Ranking[]
}

const RankingList = (props: Props) => (
  <>
    <h1>Ranking</h1>
    <ul>
      {props.rankings.map(r => (
        <li key={r.userId}>
          {r.userId}: {r.amount} ({r.languageCode})
        </li>
      ))}
    </ul>
  </>
)

export default RankingList
