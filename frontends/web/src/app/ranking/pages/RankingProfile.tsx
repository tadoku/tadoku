import React from 'react'
import { useSelector } from 'react-redux'
import styled from 'styled-components'
import ErrorPage from 'next/error'

import { ContestLog, Ranking } from '../interfaces'
import RankingApi from '../api'
import ContestApi from '../../contest/api'
import ReadingActivityGraph from '../components/graphs/ReadingActivityGraph'
import MediaDistributionGraph from '../components/graphs/MediaDistributionGraph'
import UpdatesOverview from '../components/UpdatesOverview'
import { rankingsToRegistrationOverview } from '../transform/ranking'
import { Contest } from '../../contest/interfaces'
import { useCachedApiState, isReady } from '../../cache'
import { contestSerializer } from '../../contest/transform'
import { optionalizeSerializer } from '../../transform'
import { PageTitle, ButtonLink, SubHeading } from '@app/ui/components'
import { RootState } from '../../store'
import { contestLogCollectionSerializer } from '../transform/contest-log'
import { rankingCollectionSerializer } from '../transform/ranking'
import Constants from '@app/ui/Constants'
import ScoreList from '../components/ScoreList'
import media from 'styled-media-query'
import SubmitPagesButton from '../components/SubmitPagesButton'
import { rankingRegistrationMapper } from '../transform/ranking-registration'

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
  const rankingRegistration = useSelector((state: RootState) =>
    rankingRegistrationMapper.optional.fromRaw(state.ranking.rawRegistration),
  )

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
          <ActionContainer>
            <DownloadButton
              href={dataUrl}
              download={`tadoku-contest-${contestId}-data.json`}
              icon="file-download"
            >
              Export data
            </DownloadButton>
            <SubmitPagesButton
              registration={rankingRegistration}
              refreshRanking={refreshRanking}
            />
          </ActionContainer>
        )}
      </HeaderContainer>
      <ScoreList registrationOverview={registrationOverview} />
      <GraphContainer>
        <ReadingActivityGraphContainer>
          <GraphHeading>Reading Activity</GraphHeading>
          <ReadingActivityGraph
            logs={logs}
            contest={contest}
            effectCount={effectCount}
          />
        </ReadingActivityGraphContainer>
        <MediaDistributionGraphContainer>
          <GraphHeading>Media distribution</GraphHeading>
          <MediaDistributionGraph logs={logs} effectCount={effectCount} />
        </MediaDistributionGraphContainer>
      </GraphContainer>
      <UpdatesOverview
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
  padding: 30px 30px 21px;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  box-sizing: border-box;
  background: ${Constants.colors.light};

  .rv-discrete-color-legend {
    overflow-y: hidden;

    .rv-discrete-color-legend-item {
      padding: 9px 4px;
    }
  }
`

const ReadingActivityGraphContainer = styled(LargeCard)`
  flex: 1 1 0;
  margin-right: 30px;
  max-width: calc(100% - 260px - 30px);
  ${media.lessThan('large')`max-width: 100%; margin-right: 0;`}
  ${media.lessThan('small')`display: none;`}
`
const MediaDistributionGraphContainer = styled(LargeCard)`
  ${media.lessThan('large')`display: none;`}
`

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

  ${media.lessThan('medium')`
    padding-bottom: 10px;
    margin-bottom: 20px
    flex-direction: column;
  `}
`

const RoundDescription = styled(SubHeading)`
  margin-top: 10px;
`

const GraphHeading = styled(SubHeading)`
  margin-bottom: 15px;
  margin-top: 0;
`

const ActionContainer = styled.div`
  display: flex;

  ${media.lessThan('medium')`
    width: 100%;
    button { margin: 0; flex: 1; }
  `}
`

const DownloadButton = styled(ButtonLink)`
  ${media.lessThan('medium')`display: none;`}
`
