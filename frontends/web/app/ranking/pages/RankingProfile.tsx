import React from 'react'
import Layout from '../../ui/components/Layout'
import ErrorPage from 'next/error'
import { ContestLog, RankingRegistrationOverview } from '../interfaces'
import RankingApi from '../api'
import ContestApi from '../../contest/api'
import ContestLogsByDayGraph from '../components/graphs/ContestLogsByDayGraph'
import ContestLogsOverview from '../components/ContestLogsOverview'
import {
  rankingsToRegistrationOverview,
  amountToPages,
  pagesLabel,
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
  })

  const { data: registration, status: statusRegistration } = useCachedApiState<
    RankingRegistrationOverview | undefined
  >({
    cacheKey: `ranking_registration?contest_id=${contestId}&user_id=${userId}`,
    defaultValue: undefined,
    fetchData: () => {
      return new Promise(async resolve => {
        const result = await RankingApi.getRankingsRegistration(
          contestId,
          userId,
        )

        resolve(result ? rankingsToRegistrationOverview(result) : undefined)
      })
    },
    dependencies: [contestId, userId, effectCount],
  })

  if (!isReady([statusContest, statusLogs, statusRegistration])) {
    return <Layout>Loading...</Layout>
  }

  if (!registration || !contest) {
    return <ErrorPage statusCode={500} />
  }

  if (logs.length == 0) {
    return (
      <Layout title={registration.userDisplayName}>
        <p>
          Nothing to see here! {registration.userDisplayName} hasn't logged any
          updates for this round yet, please check again later.
        </p>
      </Layout>
    )
  }

  return (
    <Layout title={registration.userDisplayName}>
      <Cards>
        <Card>
          <CardContent>{contest.description}</CardContent>
          <CardLabel>Round</CardLabel>
        </Card>
        {registration.registrations.map(r => (
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
            registration={registration}
            refreshData={refreshRanking}
          />
        </LargeCard>
      </Cards>
    </Layout>
  )
}

export default RankingProfile
