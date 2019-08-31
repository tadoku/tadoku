import React, { useState, useEffect } from 'react'
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
import { isContestActive } from '../domain'

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
    cacheKey: `ranking_overview?i=1&contest_id=${contest.id}`,
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
    isContestActive(contest) &&
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
        Tadoku&apos;s first <del>official</del> round will start September 1st
        until September 30th! Meanwhile we&apos;ll be working on improving the
        platform and adding missing features.
      </p>
      <p>
        I&apos;ve been a bit busy and haven&apos;t had the time to run a new
        test round, so it&apos;ll be a second test round. Sorry about that!
      </p>

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
    return <p>Contest has ended.</p>
  }

  return (
    <p>
      {days} day{days !== 1 && 's'} {hours} hour{hours !== 1 && 's'} {minutes}{' '}
      minute{minutes !== 1 && 's'} {seconds} second{seconds !== 1 && 's'}{' '}
      remaining
    </p>
  )
}
