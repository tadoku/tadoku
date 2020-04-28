import React, { SFC } from 'react'
import Link from 'next/link'
import styled from 'styled-components'
import ContentLoader from 'react-content-loader'

import { Ranking, RankingWithRank } from '../interfaces'
import { formatScore, calculateLeaderboard } from '../transform/graph'
import Constants from '../../ui/Constants'
import media from 'styled-media-query'

interface Props {
  rankings: Ranking[]
  loading: boolean
}

const RankingList = (props: Props) => {
  if (props.loading) {
    const rows = [...Array(5)]

    return (
      <Table>
        <Heading />
        <tbody>
          {rows.map((_, i) => (
            <RankingRowSkeleton key={i} />
          ))}
        </tbody>
      </Table>
    )
  }

  const leaderboard = calculateLeaderboard(props.rankings)

  return (
    <Table>
      <Heading />
      <tbody>
        {leaderboard.map(row => (
          <RankingRow {...row} key={row.data.userId} />
        ))}
      </tbody>
    </Table>
  )
}

export default RankingList

const RankingRow = ({ rank, tied, data: rankingData }: RankingWithRank) => (
  <Row>
    <RankCell>
      <RowLink ranking={rankingData}>
        {tied && <span title="Tied">T</span>}
        {rank}
      </RowLink>
    </RankCell>
    <NicknameCell>
      <RowLink ranking={rankingData}>{rankingData.userDisplayName}</RowLink>
    </NicknameCell>
    <ScoreCell>
      <RowLink ranking={rankingData}>{formatScore(rankingData.amount)}</RowLink>
    </ScoreCell>
  </Row>
)

const RowLink: SFC<{ ranking: Ranking }> = ({ ranking, children }) => (
  <Link
    href="/contest-profile/[contest_id]/[user_id]"
    as={`/contest-profile/${ranking.contestId}/${ranking.userId}`}
    passHref
  >
    <RowAnchor href="">{children}</RowAnchor>
  </Link>
)

const Heading = () => (
  <thead>
    <TableHeading>
      <RankHeading>Rank</RankHeading>
      <NicknameHeading>Nickname</NicknameHeading>
      <ScoreHeading>Score</ScoreHeading>
    </TableHeading>
  </thead>
)

const Table = styled.table`
  padding: 0;
  width: 100%;
  background: ${Constants.colors.light};
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  border-collapse: collapse;
`

const TableHeading = styled.tr`
  height: 55px;
  font-size: 16px;
  font-weight: bold;
  text-transform: uppercase;
  color: ${Constants.colors.nonFocusText};
`

const RankHeading = styled.td`
  width: 80px;
  padding: 0 30px;
  box-sizing: border-box;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.2)};

  ${media.lessThan('large')`
    padding: 0 20px;
  `}
`

const NicknameHeading = styled.td`
  width: 100%;
  padding: 0 30px;
  box-sizing: border-box;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.2)};

  ${media.lessThan('large')`
    padding: 0 20px;
  `}
`

const ScoreHeading = styled.td`
  min-width: 100px;
  max-width: 150px;
  padding: 0 30px;
  text-align: right;
  box-sizing: border-box;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.2)};

  ${media.lessThan('large')`
    padding: 0 20px;
  `}
`

const Row = styled.tr`
  height: 55px;
  padding: 0;
  font-size: 20px;
  font-weight: bold;
  transition: background 0.1s ease;

  &:nth-child(2n + 1) {
    background-color: ${Constants.colors.nonFocusTextWithAlpha(0.05)};
  }

  &:hover,
  &:active,
  &:focus {
    background: ${Constants.colors.primary};

    a {
      color: ${Constants.colors.light};
      transition: none;
    }
  }
`

const RowAnchor = styled.a`
  display: block;
  padding: 0;
  height: 55px;
  line-height: 55px;

  &:hover,
  &:active,
  &:focus {
    color: inherit;
  }
`

const RankCell = styled.td`
  text-align: center;
  height: 55px;
  padding: 0 30px;

  ${media.lessThan('large')`
    padding: 0 20px;
  `}
`

const NicknameCell = styled.td`
  height: 55px;
  padding: 0 30px;

  ${media.lessThan('large')`
    padding: 0 20px;
  `}
`

const ScoreCell = styled.td`
  text-align: right;
  height: 55px;
  padding: 0 30px;

  ${media.lessThan('large')`
    padding: 0 20px;
  `}
`

const RankingRowSkeleton = () => (
  <SkeletonRow>
    <RankCell>
      <Skeleton />
    </RankCell>
    <NicknameCell>
      <Skeleton />
    </NicknameCell>
    <ScoreCell>
      <Skeleton />
    </ScoreCell>
  </SkeletonRow>
)

const SkeletonRow = styled.tr`
  height: 55px;
  padding: 0;
`

const Skeleton = () => (
  <ContentLoader
    speed={2}
    style={{ width: '100%', height: '25px', borderRadius: '2px' }}
    height={25}
  >
    <rect x="0" y="0" rx="0" ry="0" width="100%" height="25" />
  </ContentLoader>
)
