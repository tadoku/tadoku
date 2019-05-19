import React from 'react'
import Link from 'next/link'
import { Ranking } from '../interfaces'
import styled from 'styled-components'

const List = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0 auto;
`

const Row = styled.li`
  margin: 20px 0;
  padding: 20px 30px;
  border-radius: 2px;
  box-shadow: 4px 5px 15px 1px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
`

const Rank = styled.div`
  width: 50px;
  font-size: 30px;
`

const Name = styled.div`
  flex: 1;
  font-size: 20px;

  a {
    display: block;
  }
`

const Pages = styled.div`
  font-size: 25px;

  span {
    font-size: 20px;
  }
`

interface Props {
  rankings: Ranking[]
}

const RankingList = (props: Props) => (
  <>
    <h1>Ranking</h1>
    <List>
      {props.rankings.map((r, rank) => (
        <Row key={r.userId}>
          <Rank>{rank + 1}</Rank>
          <Name>
            <Link
              as={`/contest/1/rankings/${r.userId}`}
              href={`/ranking-details?contest_id=1&user_id=${r.userId}`}
            >
              <a href="">{r.userDisplayName}</a>
            </Link>
          </Name>
          <Pages>
            {Math.round(r.amount * 10) / 10}
            <span> pages</span>
          </Pages>
        </Row>
      ))}
    </List>
  </>
)

export default RankingList
