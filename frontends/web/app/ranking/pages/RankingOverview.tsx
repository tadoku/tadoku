import React, { useEffect, useState } from 'react'
import Layout from '../../ui/components/Layout'
import { Ranking, RankingRegistration } from '../../ranking/interfaces'
import RankingList from '../../ranking/components/List'
import RankingApi from '../../ranking/api'
import { Contest } from '../../contest/interfaces'
import { Button } from '../../ui/components'
import styled from 'styled-components'
import { User } from '../../session/interfaces'
import JoinContestModal from '../components/modals/JoinContestModal'

interface Props {
  contest: Contest | undefined
  registration: RankingRegistration | undefined
  user: User | undefined
  effectCount: number
  refreshRegistration: () => void
}

const RankingOverview = ({
  contest,
  registration,
  user,
  effectCount,
  refreshRegistration,
}: Props) => {
  const [rankings, setRankings] = useState([] as Ranking[])
  const [joinModalOpen, setJoinModalOpen] = useState(false)

  useEffect(() => {
    if (!contest) {
      return
    }

    const update = async () => {
      const payload = await RankingApi.get(contest.id)
      setRankings(payload)
    }
    update()
  }, [contest, effectCount])

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
    <Layout title="Ranking">
      <Container>
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
                refreshRegistration()
              }}
              onCancel={() => setJoinModalOpen(false)}
            />
          </>
        )}
      </Container>
      <p>
        We'll be running a test round from June 15th until June 30th UTC. All
        existing data will be wiped after this. Registrations are open now.
      </p>
      <RankingList rankings={rankings} />
    </Layout>
  )
}

export default RankingOverview

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`
