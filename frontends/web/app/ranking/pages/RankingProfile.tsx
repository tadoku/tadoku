import React from 'react'
import ErrorPage from 'next/error'
import { ContestLog, Ranking } from '../interfaces'
import RankingApi from '../api'
import ContestApi from '../../contest/api'
import ContestLogsByDayGraph from '../components/graphs/ContestLogsByDayGraph'
import ContestLogsOverview from '../components/ContestLogsOverview'
import {
  rankingsToRegistrationOverview,
  amountToPages,
  pagesLabel,
  ContestLogsSerializer,
  RankingsSerializer,
} from '../transform'
import { Contest } from '../../contest/interfaces'
import Cards, {
  Card,
  CardLabel,
  CardContent,
  LargeCard,
} from '../../ui/components/Cards'
import { useCachedApiState, isReady } from '../../cache'
import { ContestSerializer } from '../../contest/transform'
import { OptionalizeSerializer } from '../../transform'
import { PageTitle } from '../../ui/components'

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
  const { data: contest, status: statusContest } = useCachedApiState<
    Contest | undefined
  >({
    cacheKey: `contest?id=${contestId}`,
    defaultValue: undefined,
    fetchData: () => {
      return ContestApi.get(contestId)
    },
    dependencies: [contestId],
    serializer: OptionalizeSerializer(ContestSerializer),
  })

  const { data: logs, status: statusLogs } = useCachedApiState<ContestLog[]>({
    cacheKey: `contest_logs?contest_id=${contestId}&user_id=${userId}`,
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
    serializer: ContestLogsSerializer,
  })

  const { data: registration, status: statusRegistration } = useCachedApiState<
    Ranking[]
  >({
    cacheKey: `ranking_profile_registration?contest_id=${contestId}&user_id=${userId}`,
    defaultValue: [],
    fetchData: () => {
      return RankingApi.getRankingsRegistration(contestId, userId)
    },
    dependencies: [contestId, userId, effectCount],
    serializer: RankingsSerializer,
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
        <p>
          Nothing to see here! {registrationOverview.userDisplayName} hasn't
          logged any updates for this round yet, please check again later.
        </p>
      </>
    )
  }

  return (
    <>
      <PageTitle>{registrationOverview.userDisplayName}</PageTitle>
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
          <ContestLogsOverview
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
