import React, { useState } from 'react'
import Layout from '../../ui/components/Layout'
import { Ranking, RankingRegistration } from '../../ranking/interfaces'
import RankingList from '../../ranking/components/List'
import RankingApi from '../../ranking/api'
import { Contest } from '../../contest/interfaces'
import { Button, PageTitle } from '../../ui/components'
import styled from 'styled-components'
import { User } from '../../session/interfaces'
import JoinContestModal from '../components/modals/JoinContestModal'
import { useCachedApiState, ApiFetchStatus } from '../../cache'
import { RankingsSerializer } from '../transform'

interface Props {
  contest: Contest
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
  const [joinModalOpen, setJoinModalOpen] = useState(false)

  const { data: rankings, status } = useCachedApiState<Ranking[]>({
    cacheKey: `ranking_overview?contest_id=${contest.id}`,
    defaultValue: [],
    fetchData: () => {
      if (!contest) {
        return new Promise<Ranking[]>(resolve => resolve([]))
      }

      return RankingApi.get(contest.id)
    },
    dependencies: [contest, effectCount],
    serializer: RankingsSerializer,
  })

  // @TODO: extract this business logic
  const canJoin =
    user &&
    contest &&
    contest.open &&
    contest.end > new Date() &&
    ((registration && registration.contestId !== contest.id) || !registration)

  return (
    <>
      <PageTitle>Ranking</PageTitle>
      <Container>
        {canJoin && contest && (
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
      <RankingList
        rankings={rankings}
        loading={status === ApiFetchStatus.Loading}
      />
    </>
  )
}

export default RankingOverview

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`
