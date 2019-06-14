import React from 'react'
import Link from 'next/link'
import ContentLoader from 'react-content-loader'
import { Ranking } from '../interfaces'
import styled from 'styled-components'
import { amountToPages } from '../transform'

interface Props {
  rankings: Ranking[]
  loading: boolean
}

const RankingList = (props: Props) => {
  if (props.loading) {
    const rows = [...Array(5)]

    return (
      <List>
        {rows.map((_, i) => (
          <RankingRowSkeleton key={i} rank={i + 1} />
        ))}
      </List>
    )
  }

  return (
    <List>
      {props.rankings.map((r, rank) => (
        <RankingRow rank={rank} data={r} key={r.userId} />
      ))}
    </List>
  )
}

export default RankingList

const RankingRow = ({ rank, data }: { rank: number; data: Ranking }) => (
  <Row>
    <Link
      as={`/contest/1/rankings/${data.userId}`}
      href={`/ranking-details?contest_id=1&user_id=${data.userId}`}
    >
      <RowLink href="">
        <Rank>{rank + 1}</Rank>
        <Name>{data.userDisplayName}</Name>
        <Pages>
          {amountToPages(data.amount)}
          <PagesLabel> pages</PagesLabel>
        </Pages>
      </RowLink>
    </Link>
  </Row>
)

const RankingRowSkeleton = ({ rank }: { rank: number }) => (
  <Row>
    <RowLink href="">
      <Rank>{rank}</Rank>
      <Name>
        <Skeleton />
      </Name>
      <PagesLoading>
        <Skeleton />
      </PagesLoading>
    </RowLink>
  </Row>
)

const Skeleton = () => (
  <ContentLoader
    speed={2}
    style={{ width: '100%', height: '25px' }}
    height={25}
  >
    <rect x="0" y="0" rx="0" ry="0" width="100%" height="25" />
  </ContentLoader>
)

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
  margin-left: 30px;
  font-size: 25px;
`

const PagesLoading = styled.div`
  margin-left: 30px;
  min-width: 50px;
  max-width: 100px;
  font-size: 25px;
`

const PagesLabel = styled.span`
  font-size: 20px;
`
