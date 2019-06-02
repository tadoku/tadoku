import React, { useEffect, useState } from 'react'
import Layout from '../../ui/components/Layout'
import { Ranking, RankingRegistration } from '../../ranking/interfaces'
import RankingList from '../../ranking/components/List'
import RankingApi from '../../ranking/api'
import { connect } from 'react-redux'
import { State } from '../../store'
import { Contest } from '../../contest/interfaces'
import { Button } from '../../ui/components'
import styled from 'styled-components'
import { User } from '../../user/interfaces'
import JoinContestModal from '../../ranking/components/JoinContestModal'

interface Props {
  contest: Contest | undefined
  registration: RankingRegistration | undefined
  user: User | undefined
}

const RankingOverview = ({ contest, registration, user }: Props) => {
  const [rankings, setRankings] = useState([] as Ranking[])
  const [joinModalOpen, setJoinModalOpen] = useState(false)
  const [updateCount, setUpdateCount] = useState(0)

  useEffect(() => {
    if (!contest) {
      return
    }

    const update = async () => {
      const payload = await RankingApi.get(contest.id)
      setRankings(payload)
    }
    update()
  }, [contest, updateCount])

  if (!rankings || !contest) {
    return <Layout>No ranking found.</Layout>
  }

  // @TODO: extract this business logic
  const canJoin =
    user &&
    contest &&
    contest.open &&
    contest.end > new Date() &&
    ((registration && registration.contestId !== contest.id) || !registration)

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
              contest={contest}
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
  contest: state.contest.latestContest,
  registration: state.ranking.registration,
  user: state.session.user,
})

export default connect(mapStateToProps)(RankingOverview)

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`
