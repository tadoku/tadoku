import React, { useState, useEffect } from 'react'
import Layout from '../../ui/components/Layout'
import ErrorPage from 'next/error'
import { ContestLog, RankingRegistrationOverview } from '../interfaces'
import RankingApi from '../api'
import ContestApi from '../../contest/api'
import ContestLogsByDayGraph from '../components/graphs/ContestLogsByDayGraph'
import ContestLogsList from '../components/ContestLogsList'
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

interface Props {
  contestId: number | undefined
  userId: number | undefined
  effectCount: number
  refreshRanking: () => void
}

const RankingProfile = ({
  contestId,
  userId,
  effectCount,
  refreshRanking,
}: Props) => {
  const [loaded, setLoaded] = useState(false)
  const [logs, setLogs] = useState([] as ContestLog[])
  const [contest, setContest] = useState(undefined as Contest | undefined)
  const [registration, setRegistration] = useState(undefined as
    | RankingRegistrationOverview
    | undefined)

  useEffect(() => {
    if (!contestId || !userId) {
      return
    }

    const getLogs = async () => {
      const [contest, logs, registration] = await Promise.all([
        ContestApi.get(contestId),
        RankingApi.getLogsFor(contestId, userId),
        RankingApi.getRankingsRegistration(contestId, userId),
      ])

      setContest(contest)
      setLogs(
        logs.sort((a, b) => {
          if (a.date > b.date) {
            return 1
          }
          if (a.date < b.date) {
            return -1
          }
          return 0
        }),
      )
      setRegistration(rankingsToRegistrationOverview(registration))
      setLoaded(true)
    }

    getLogs()
  }, [contestId, userId, effectCount])

  if (!contestId || !userId) {
    return <ErrorPage statusCode={404} />
  }

  if (!loaded) {
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

  const today = new Date()
  today.setHours(0, 0, 0, 0)

  const contestForGraphs = {
    ...contest,
    end: contest.end > today ? today : contest.end,
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
          <ContestLogsByDayGraph logs={logs} contest={contestForGraphs} />
        </LargeCard>
        <LargeCard>
          <ContestLogsList
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
