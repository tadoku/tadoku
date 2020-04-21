import React, { useState, useEffect } from 'react'
import { Ranking, RankingRegistration } from '../interfaces'
import RankingList from '../components/List'
import RankingApi from '../api'
import { Contest } from '../../contest/interfaces'
import { Button, PageTitle } from '../../ui/components'
import styled from 'styled-components'
import { User } from '../../session/interfaces'
import JoinContestModal from '../components/modals/JoinContestModal'
import { useCachedApiState, ApiFetchStatus } from '../../cache'
import { RankingsSerializer } from '../transform/ranking'
import { isContestActive } from '../domain'
import SubmitPagesButton from '../components/SubmitPagesButton'

interface Props {
  contest: Contest
  registration: RankingRegistration | undefined
  user: User | undefined
  effectCount: number
  refreshRanking: () => void
}

const RankingOverview = ({
  contest,
  registration,
  user,
  effectCount,
  refreshRanking,
}: Props) => {
  const [joinModalOpen, setJoinModalOpen] = useState(false)

  const { data: rankings, status } = useCachedApiState<Ranking[]>({
    cacheKey: `ranking_overview?i=1&contest_id=${contest.id}`,
    defaultValue: [],
    fetchData: () => {
      if (!contest) {
        return new Promise<Ranking[]>(resolve => resolve([]))
      }

      return RankingApi.get(contest.id)
    },
    dependencies: [contest?.id, effectCount],
    serializer: RankingsSerializer,
  })

  // @TODO: extract this business logic
  const isActive = user && contest && isContestActive(contest)
  const isRegistered = registration?.contestId === contest.id
  const canJoin = isActive && !isRegistered

  return (
    <>
      <Container>
        <PageTitle>Ranking</PageTitle>
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
                refreshRanking()
              }}
              onCancel={() => setJoinModalOpen(false)}
            />
          </>
        )}
        {isRegistered && (
          <SubmitPagesButton
            registration={registration}
            refreshRanking={refreshRanking}
          />
        )}
      </Container>
      {new Date() >= contest.start && <RemainingUntil date={contest.end} />}
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

// @TODO: Make this a proper component, too quick and dirty atm
const RemainingUntil = ({ date }: { date: Date }) => {
  const [currentDate, setCurrentDate] = useState(() => new Date())

  useEffect(() => {
    const id = setInterval(() => {
      setCurrentDate(new Date())
    }, 1000)

    return () => clearInterval(id)
  }, [])

  if (!date.getTime) {
    return null
  }

  const t = date.getTime() - currentDate.getTime()
  const seconds = Math.floor((t / 1000) % 60)
  const minutes = Math.floor((t / 1000 / 60) % 60)
  const hours = Math.floor((t / (1000 * 60 * 60)) % 24)
  const days = Math.floor(t / (1000 * 60 * 60 * 24))

  if (t <= 0) {
    return <Notes>Contest has ended.</Notes>
  }

  return (
    <Notes>
      {days} day{days !== 1 && 's'} {hours} hour{hours !== 1 && 's'} {minutes}{' '}
      minute{minutes !== 1 && 's'} {seconds} second{seconds !== 1 && 's'}{' '}
      remaining
    </Notes>
  )
}

const Notes = styled.p`
  padding: 0 0 20px 0;
`
