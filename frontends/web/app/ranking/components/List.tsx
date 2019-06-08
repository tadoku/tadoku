import React from 'react'
import Link from 'next/link'
import { Ranking } from '../interfaces'
import styled from 'styled-components'
import { amountToPages } from '../transform'

interface Props {
  rankings: Ranking[]
}

const RankingList = (props: Props) => (
  <List>
    {props.rankings.map((r, rank) => (
      <Row key={r.userId}>
        <Link
          as={`/contest/1/rankings/${r.userId}`}
          href={`/ranking-details?contest_id=1&user_id=${r.userId}`}
        >
          <RowLink href="">
            <Rank>{rank + 1}</Rank>
            <Name>{r.userDisplayName}</Name>
            <Pages>
              {amountToPages(r.amount)}
              <span> pages</span>
            </Pages>
          </RowLink>
        </Link>
      </Row>
    ))}
  </List>
)

export default RankingList

const List = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0 auto;
`

const Row = styled.li`
  margin: 20px 0;
`

const RowLink = styled.a`
  padding: 20px 30px;
  border-radius: 2px;
  box-shadow: 4px 5px 15px 1px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  transition: all 0.2s ease;

  &:hover,
  &:focus,
  &:active {
    background-color: rgba(0, 0, 0, 0.02);
    box-shadow: 4px 5px 15px 1px rgba(0, 0, 0, 0.12);
  }
`

const Rank = styled.div`
  width: 50px;
  font-size: 30px;
`

const Name = styled.div`
  flex: 1;
  font-size: 20px;
`

const Pages = styled.div`
  font-size: 25px;

  span {
    font-size: 20px;
  }
`
