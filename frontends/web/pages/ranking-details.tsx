import React from 'react'
import Layout from '../app/ui/components/Layout'
import ErrorPage from 'next/error'
import { ExpressNextContext } from '../app/interfaces'

interface Props {
  contestId: number | undefined
  userId: number | undefined
}

const RankingDetails = ({ contestId, userId }: Props) => {
  if (!contestId || !userId) {
    return <ErrorPage statusCode={404} />
  }

  return <Layout>This will show the details of a ranking</Layout>
}

RankingDetails.getInitialProps = async ({ req }: ExpressNextContext) => {
  if (!req || !req.params) {
    return {}
  }

  const { contest_id, user_id } = req.params

  return {
    contestId: parseInt(contest_id),
    userId: parseInt(user_id),
  }
}

export default RankingDetails
