import React, { useState, useEffect } from 'react'
import Layout from '../app/ui/components/Layout'
import ErrorPage from 'next/error'
import { ExpressNextContext } from '../app/interfaces'
import { ContestLog } from '../app/ranking/interfaces'
import RankingApi from '../app/ranking/api'
import ContestLogsByDayGraph from '../app/ranking/components/ContestLogsByDayGraph'

interface Props {
  contestId: number | undefined
  userId: number | undefined
}

const RankingDetails = ({ contestId, userId }: Props) => {
  const [logs, setLogs] = useState([] as ContestLog[])

  useEffect(() => {
    if (!contestId || !userId) {
      return
    }

    const getLogs = async () => {
      const payload = await RankingApi.getLogsFor(contestId, userId)
      setLogs(payload)
    }

    getLogs()
  }, [contestId, userId])

  if (!contestId || !userId) {
    return <ErrorPage statusCode={404} />
  }

  return (
    <Layout>
      <ContestLogsByDayGraph logs={logs} />
    </Layout>
  )
}

RankingDetails.getInitialProps = async ({ req, query }: ExpressNextContext) => {
  if (req && req.params) {
    const { contest_id, user_id } = req.params

    return {
      contestId: parseInt(contest_id),
      userId: parseInt(user_id),
    }
  }

  if (query.contest_id && query.user_id) {
    const { contest_id, user_id } = query

    return {
      contestId: parseInt(contest_id as string),
      userId: parseInt(user_id as string),
    }
  }

  return {}
}

export default RankingDetails
