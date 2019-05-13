import React from 'react'
import Layout from '../app/ui/components/Layout'
import { Ranking } from '../app/ranking/interfaces'
import RankingList from '../app/ranking/components/List'
import RankingApi from '../app/ranking/api'

interface Props {
  rankings: Ranking[]
}

const Home = (props: Props) => {
  return (
    <Layout>
      <RankingList rankings={props.rankings} />
    </Layout>
  )
}

Home.getInitialProps = async (_: any) => {
  const rankings = await RankingApi.get(1)

  if (rankings.length > 0) {
    return { rankings }
  }

  return { rankings: [] }
}

export default Home
