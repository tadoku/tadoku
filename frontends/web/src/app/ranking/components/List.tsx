import React, { SFC } from 'react'
import Link from 'next/link'
import styled from 'styled-components'

import { Ranking, RankingWithRank } from '../interfaces'
import { amountToPages, calculateLeaderboard } from '../transform/graph'
import Constants from '../../ui/Constants'

interface Props {
  rankings: Ranking[]
  loading: boolean
}

const RankingList = (props: Props) => {
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
      <RowLink ranking={rankingData}>
        {amountToPages(rankingData.amount)}
      </RowLink>
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
  padding: 0 30px 0 60px;
  box-sizing: border-box;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.2)};
`

const NicknameHeading = styled.td`
  width: 100%;
  padding: 0 30px;
  box-sizing: border-box;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.2)};
`

const ScoreHeading = styled.td`
  min-width: 100px;
  max-width: 150px;
  padding: 0 60px 0 30px;
  text-align: right;
  box-sizing: border-box;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.2)};
`

const Row = styled.tr`
  height: 55px;
  padding: 0;
  font-size: 20px;
  font-weight: bold;
  transition: background 0.1s ease;

  &:nth-child(2n) {
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
  padding-left: 60px;
  padding-right: 30px;
`

const NicknameCell = styled.td`
  height: 55px;
  padding: 0 30px;
`

const ScoreCell = styled.td`
  text-align: right;
  height: 55px;
  padding-right: 60px;
`
