import React from 'react'
import { useSelector } from 'react-redux'
import styled from 'styled-components'
import ErrorPage from 'next/error'

import { ContestLog, Ranking } from '../interfaces'
import RankingApi from '../api'
import ContestApi from '../../contest/api'
import ContestLogsByDayGraph from '../components/graphs/ReadingActivityGraph'
import ContestLogsByMediumGraph from '../components/graphs/MediaDistributionGraph'
import ContestLogsOverview from '../components/ContestLogsOverview'
import { rankingsToRegistrationOverview } from '../transform/ranking'
import { Contest } from '../../contest/interfaces'
import { useCachedApiState, isReady } from '../../cache'
import { contestSerializer } from '../../contest/transform'
import { optionalizeSerializer } from '../../transform'
import { PageTitle, ButtonLink, SubHeading } from '../../ui/components'
import { RootState } from '../../store'
import { contestLogCollectionSerializer } from '../transform/contest-log'
import { rankingCollectionSerializer } from '../transform/ranking'
import Constants from '../../ui/Constants'
import ScoreList from '../components/ScoreList'
import media from 'styled-media-query'

interface Props {
  contestId: number
  userId: number
  effectCount: number
  refreshRanking: () => void
}

const RankingProfile = ({
  contestId,
  userId,
  effectCount,
  refreshRanking,
}: Props) => {
  const signedInUser = useSelector((state: RootState) => state.session.user)
  const { data: contest, status: statusContest } = useCachedApiState<
    Contest | undefined
  >({
    cacheKey: `contest?i=1&id=${contestId}`,
    defaultValue: undefined,
    fetchData: () => {
      return ContestApi.get(contestId)
    },
    dependencies: [contestId],
    serializer: optionalizeSerializer(contestSerializer),
  })

  const { data: logs, status: statusLogs } = useCachedApiState<ContestLog[]>({
    cacheKey: `contest_logs?i=1&contest_id=${contestId}&user_id=${userId}`,
    defaultValue: [],
    fetchData: () => {
      return new Promise(async resolve => {
        const result = (await RankingApi.getLogsFor(contestId, userId)).sort(
          (a, b) => {
            if (a.date > b.date) {
              return -1
            }
            if (a.date < b.date) {
              return 1
            }
            return 0
          },
        )

        resolve(result)
      })
    },
    dependencies: [contestId, userId, effectCount],
    serializer: contestLogCollectionSerializer,
  })

  const { data: registration, status: statusRegistration } = useCachedApiState<
    Ranking[]
  >({
    cacheKey: `ranking_registration?i=1&contest_id=${contestId}&user_id=${userId}`,
    defaultValue: [],
    fetchData: () => {
      return RankingApi.getRankingsRegistration(contestId, userId)
    },
    dependencies: [contestId, userId, effectCount],
    serializer: rankingCollectionSerializer,
  })

  if (!isReady([statusContest, statusLogs, statusRegistration])) {
    return <p>Loading...</p>
  }

  const registrationOverview = rankingsToRegistrationOverview(registration)

  if (!registrationOverview || !contest) {
    return <ErrorPage statusCode={500} />
  }

  if (logs.length == 0) {
    return (
      <>
        <PageTitle>{registrationOverview.userDisplayName}</PageTitle>
        <RoundDescription>{contest.description}</RoundDescription>
        <p>
          Nothing to see here! {registrationOverview.userDisplayName}{' '}
          hasn&apos;t logged any updates for this round yet, please check again
          later.
        </p>
      </>
    )
  }

  const dataUrl = `data:text/json;charset=utf-8,${encodeURIComponent(
    JSON.stringify(logs, null, 2),
  )}`

  return (
    <>
      <HeaderContainer>
        <div>
          <PageTitle>{registrationOverview.userDisplayName}</PageTitle>
          <RoundDescription>{contest.description}</RoundDescription>
        </div>
        {signedInUser && userId === signedInUser.id && (
          <ButtonLink
            href={dataUrl}
            download={`tadoku-contest-${contestId}-data.json`}
            icon="file-download"
          >
            Export data
          </ButtonLink>
        )}
      </HeaderContainer>
      <ScoreList registrationOverview={registrationOverview} />
      <GraphContainer>
        <OverallGraph>
          <GraphHeading>Reading Activity</GraphHeading>
          <ContestLogsByDayGraph logs={logs} contest={contest} />
        </OverallGraph>
        <MediaGraph>
          <GraphHeading>Media distribution</GraphHeading>
          <ContestLogsByMediumGraph logs={logs} />
        </MediaGraph>
      </GraphContainer>
      <ContestLogsOverview
        contest={contest}
        logs={logs}
        registration={registrationOverview}
        refreshData={refreshRanking}
      />
    </>
  )
}

export default RankingProfile

const GraphContainer = styled.div`
  display: flex;
  justify-content: space-between;
  width: 100%;
  flex-wrap: nowrap;

  ${media.lessThan('large')`
    flex-wrap: wrap;
    padding-bottom: 30px;
  `}
`

const LargeCard = styled.div`
  margin-top: 30px;
  padding: 30px;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  box-sizing: border-box;
  background: ${Constants.colors.light};
`

const OverallGraph = styled(LargeCard)`
  flex: 1 1 0;
  margin-right: 30px;
  padding-bottom: 21px;
  max-width: calc(100% - 260px - 30px);
  ${media.lessThan('large')`max-width: 100%;`}
`
const MediaGraph = styled(LargeCard)``

const HeaderContainer = styled.div`
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 60px;
  border-bottom: 2px solid ${Constants.colors.lightGray};
  padding-bottom: 30px;

  h1 {
    margin: 0;
  }
`

const RoundDescription = styled(SubHeading)`
  margin-top: 10px;
`

const GraphHeading = styled(SubHeading)`
  margin-bottom: 15px;
  margin-top: 0;
`
