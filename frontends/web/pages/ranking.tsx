import React, { useEffect, useState } from 'react'
import Layout from '../app/ui/components/Layout'
import { Ranking, RankingRegistration } from '../app/ranking/interfaces'
import RankingList from '../app/ranking/components/List'
import RankingApi from '../app/ranking/api'
import { connect } from 'react-redux'
import { State } from '../app/store'
import { Contest } from '../app/contest/interfaces'
import { Button } from '../app/ui/components'
import styled from 'styled-components'
import { User } from '../app/user/interfaces'
import JoinContestModal from '../app/ranking/components/JoinContestModal'

interface Props {
  latestContest: Contest | undefined
  registration: RankingRegistration | undefined
  user: User | undefined
}

const Home = ({ latestContest, registration, user }: Props) => {
  const [rankings, setRankings] = useState([] as Ranking[])
  const [joinModalOpen, setJoinModalOpen] = useState(false)
  const [updateCount, setUpdateCount] = useState(0)

  useEffect(() => {
    if (!latestContest) {
      return
    }

    const update = async () => {
      const payload = await RankingApi.get(latestContest.id)
      setRankings(payload)
    }
    update()
  }, [latestContest, updateCount])

  if (!rankings || !latestContest) {
    return <Layout>No ranking found.</Layout>
  }

  // @TODO: extract this business logic
  const canJoin =
    user &&
    latestContest &&
    latestContest.open &&
    latestContest.end > new Date() &&
    ((registration && registration.contestId !== latestContest.id) ||
      !registration)

  return (
    <Layout>
      <Container>
        <h1>Ranking</h1>
        {canJoin && (
          <>
            <Button primary large onClick={() => setJoinModalOpen(true)}>
              Join contest
            </Button>
            <JoinContestModal
              contest={latestContest}
              isOpen={joinModalOpen}
              onSuccess={() => {
                setJoinModalOpen(false)
                setUpdateCount(updateCount + 1)
              }}
              onCancel={() => setJoinModalOpen(false)}
            />
          </>
        )}
      </Container>
      <RankingList rankings={rankings} />
    </Layout>
  )
}

const mapStateToProps = (state: State) => ({
  latestContest: state.contest.latestContest,
  registration: state.ranking.registration,
  user: state.session.user,
})

export default connect(mapStateToProps)(Home)

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`
