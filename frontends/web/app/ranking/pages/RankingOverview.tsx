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
  contest: Contest
  registration: RankingRegistration | undefined
  user: User | undefined
  effectCount: number
  refreshRegistration: () => void
}

enum ApiFetchStatus {
  Initialized,
  Stale,
  Loading,
  Completed,
}

const useCachedApiState = (
  cacheKey: string,
  defaultValue: any,
  fetchData: () => Promise<any>,
  dependencies: any[],
) => {
  const [status, setStatus] = useState(ApiFetchStatus.Initialized)
  const [data, setData] = useState(defaultValue)

  const reloadData = async () => {
    const cachedValue = localStorage.getItem(cacheKey)
    if (cachedValue) {
      setStatus(ApiFetchStatus.Stale)
      setData(JSON.parse(cachedValue))
    } else {
      setStatus(ApiFetchStatus.Loading)
    }

    const fetchedData = await fetchData()
    setData(fetchedData)
    setStatus(ApiFetchStatus.Completed)

    localStorage.setItem(cacheKey, JSON.stringify(fetchedData))
  }

  useEffect(() => {
    const update = async () => await reloadData()
    update()
  }, dependencies)

  return [data, status, reloadData]
}

const RankingOverview = ({
  contest,
  registration,
  user,
  effectCount,
  refreshRegistration,
}: Props) => {
  const [joinModalOpen, setJoinModalOpen] = useState(false)

  const [rankings, status] = useCachedApiState(
    `ranking_overview?contest_id=${contest.id}`,
    [] as Ranking[],
    () => {
      if (!contest) {
        return new Promise<Ranking[]>(resolve => resolve([]))
      }

      return RankingApi.get(contest.id)
    },
    [contest, effectCount],
  )

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
    </Layout>
  )
}

export default RankingOverview

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`
