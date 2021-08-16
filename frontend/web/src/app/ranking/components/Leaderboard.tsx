import React, { SFC } from 'react'
import Link from 'next/link'
import styled from 'styled-components'
import ContentLoader from 'react-content-loader'

import { Ranking, RankingWithRank } from '../interfaces'
import Constants from '@app/ui/Constants'
import media from 'styled-media-query'
import { aggregateRankingLeaderboard } from '../transform/ranking-leaderboard'
import { formatScore } from '../transform/format'

interface Props {
  rankings: Ranking[]
  loading: boolean
}

const Leaderboard = (props: Props) => {
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

  const leaderboard = aggregateRankingLeaderboard(props.rankings)

  return (
    <Table>
      <Heading />
      <tbody>
        {leaderboard.map(row => (
          <RankingRow {...row} key={row.data.userId} />
        ))}
        {leaderboard.length === 0 && <EmptyRow />}
      </tbody>
    </Table>
  )
}

export default Leaderboard

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
      <ScoreLarge>
        <RowLink ranking={rankingData}>
          {formatScore(rankingData.amount)}
        </RowLink>
      </ScoreLarge>
      <ScoreSmall>
        <RowLink ranking={rankingData}>
          {Math.floor(rankingData.amount)}
        </RowLink>
      </ScoreSmall>
    </ScoreCell>
  </Row>
)

const EmptyRow = () => (
  <BaseRow>
    <ErrorCell>No participants found</ErrorCell>
  </BaseRow>
)

// @TODO: refactor styled components with table components from ui package

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
  width: 100%;
  padding: 0;
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
  padding: 0 30px;
  box-sizing: border-box;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.2)};

  ${media.lessThan('large')`
    padding: 0 20px;
  `}

  ${media.lessThan('medium')`
    padding: 0;
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

const ErrorCell = styled.td.attrs({ colSpan: 3 })`
  padding: 50px 30px;
  text-align: center;
  color: ${Constants.colors.darkWithAlpha(0.5)};

  ${media.lessThan('large')`
    padding: 50px 20px;
  `}
`

const BaseRow = styled.tr`
  height: 55px;
  padding: 0;
  font-size: 20px;
  font-weight: bold;
`

const Row = styled(BaseRow)`
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
  position: relative;

  &:before {
    content: '&nbsp;';
    visibility: hidden;
  }

  a {
    height: 55px;
    line-height: 55px;
    position: absolute;
    top: 0;
    left: 30px;
    right: 30px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  ${media.lessThan('large')`
    padding: 0 20px;

    a {
      left: 20px;
      right: 20px;
    }
  `}

  ${media.lessThan('medium')`
    padding: 0;

    a {
      left: 0;
      right: 0;
    }
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

const ScoreSmall = styled.span`
  display: none;
  ${media.lessThan('medium')`display: block;`}
`
const ScoreLarge = styled.span`
  display: block;
  ${media.lessThan('medium')`display: none;`}
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
