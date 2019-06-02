import React, { useEffect, useState } from 'react'
import Layout from '../app/ui/components/Layout'
import { Ranking } from '../app/ranking/interfaces'
import RankingList from '../app/ranking/components/List'
import RankingApi from '../app/ranking/api'
import { connect } from 'react-redux'
import { State } from '../app/store'
import { Contest } from '../app/contest/interfaces'

interface Props {
  latestContest: Contest | undefined
}

const Home = ({ latestContest }: Props) => {
  const [rankings, setRankings] = useState([] as Ranking[])
  useEffect(() => {
    if (!latestContest) {
      return
    }

    const update = async () => {
      const payload = await RankingApi.get(latestContest.id)
      setRankings(payload)
    }
    update()
  }, [latestContest])

  if (!rankings) {
    return <Layout>No ranking found.</Layout>
  }

  return (
    <Layout>
      <h1>Ranking</h1>
      <RankingList rankings={rankings} />
    </Layout>
  )
}

const mapStateToProps = (state: State) => ({
  latestContest: state.contest.latestContest,
})

export default connect(mapStateToProps)(Home)
