import React from 'react'
import Link from 'next/link'
import ContentLoader from 'react-content-loader'
import { Ranking, RankingWithRank } from '../interfaces'
import styled from 'styled-components'
import { amountToPages, calculateLeaderboard } from '../transform'

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

  const leaderboard = calculateLeaderboard(props.rankings)

  return (
    <List>
      {leaderboard.map(row => (
        <RankingRow {...row} key={row.data.userId} />
      ))}
    </List>
  )
}

export default RankingList

const RankingRow = ({ rank, tied, data }: RankingWithRank) => (
  <Row>
    <Link
      as={`/contest/1/rankings/${data.userId}`}
      href={`/ranking-details?contest_id=1&user_id=${data.userId}`}
    >
      <RowLink href="">
        <Rank>
          {tied && <span title="Tied">T</span>}
          {rank}
        </Rank>
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
    style={{ width: '100%', height: '25px', borderRadius: '2px' }}
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
  font-size: 30px;
  margin-right: 30px;

  span {
    padding-right: 5px;
  }
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
