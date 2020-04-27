import React from 'react'
import ErrorPage from 'next/error'
import { ContestLog, Ranking } from '../interfaces'
import RankingApi from '../api'
import ContestApi from '../../contest/api'
import ContestLogsByDayGraph from '../components/graphs/ContestLogsByDayGraph'
import ContestLogsByMediumGraph from '../components/graphs/ContestLogsByMediumGraph'
import ContestLogsOverview from '../components/ContestLogsOverview'
import {
  rankingsToRegistrationOverview,
  amountToPages,
  pagesLabel,
} from '../transform/graph'
import { Contest } from '../../contest/interfaces'
import Cards, {
  Card,
  CardLabel,
  CardContent,
  LargeCard,
} from '../../ui/components/Cards'
import { useCachedApiState, isReady } from '../../cache'
import { contestSerializer } from '../../contest/transform'
import { optionalizeSerializer } from '../../transform'
import { PageTitle, ButtonLink, SubHeading } from '../../ui/components'
import { useSelector } from 'react-redux'
import { RootState } from '../../store'
import styled from 'styled-components'
import { contestLogCollectionSerializer } from '../transform/contest-log'
import { rankingCollectionSerializer } from '../transform/ranking'

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
      <Cards>
        <Card>
          <CardContent>{contest.description}</CardContent>
          <CardLabel>Round</CardLabel>
        </Card>
        {registrationOverview.registrations.map(r => (
          <Card key={r.languageCode}>
            <CardContent>{amountToPages(r.amount)}</CardContent>
            <CardLabel>{pagesLabel(r.languageCode)}</CardLabel>
          </Card>
        ))}
        <LargeCard>
          <ContestLogsByDayGraph logs={logs} contest={contest} />
        </LargeCard>
        <LargeCard>
          <ContestLogsByMediumGraph logs={logs} />
        </LargeCard>
        <LargeCard>
          <ContestLogsOverview
            contest={contest}
            logs={logs}
            registration={registrationOverview}
            refreshData={refreshRanking}
          />
        </LargeCard>
      </Cards>
    </>
  )
}

export default RankingProfile

const HeaderContainer = styled.div`
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 30px;

  h1 {
    margin: 0;
  }

  h2 {
    margin: 10px 0 0 0;
  }
`

const RoundDescription = styled(SubHeading)``
