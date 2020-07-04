import React, { useState } from 'react'
import { Ranking, RankingRegistration } from '../interfaces'
import Leaderboard from '../components/Leaderboard'
import RankingApi from '../api'
import { Contest } from '@app/contest/interfaces'
import { Button, PageTitle, SubHeading } from '@app/ui/components'
import styled from 'styled-components'
import { User } from '@app/session/interfaces'
import JoinContestModal from '../components/modals/JoinContestModal'
import { useCachedApiState, ApiFetchStatus } from '../../cache'
import { rankingCollectionSerializer } from '../transform/ranking'
import { isRegisteredForContest, canJoinContest } from '../domain'
import SubmitPagesButton from '../components/SubmitPagesButton'
import media from 'styled-media-query'

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
    serializer: rankingCollectionSerializer,
  })

  const isRegistered = isRegisteredForContest(registration, contest)
  const canJoin = canJoinContest(user, registration, contest)

  return (
    <>
      <Container>
        <div>
          <PageTitle>Ranking</PageTitle>
          <Description>{contest.description}</Description>
        </div>
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
      <Leaderboard
        rankings={rankings}
        loading={status === ApiFetchStatus.Loading}
      />
    </>
  )
}

export default RankingOverview

const Container = styled.div`
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 30px;

  h1 {
    margin: 0;
  }

  ${media.lessThan('small')`
    flex-direction: column;

    > button {
      width: 100%;
      box-sizing: border-box;
      margin: 10px 0;
    }
  `}
`

const Description = styled(SubHeading)`
  margin-top: 10px;
  margin-bottom: 10px;
`

// TODO: Refactor remaining time
// const RemainingUntil = ({ date }: { date: Date }) => {
//   const [currentDate, setCurrentDate] = useState(() => new Date())

//   useEffect(() => {
//     const id = setInterval(() => {
//       setCurrentDate(new Date())
//     }, 1000)

//     return () => clearInterval(id)
//   }, [])

//   if (!date.getTime) {
//     return null
//   }

//   const t = date.getTime() - currentDate.getTime()
//   const seconds = Math.floor((t / 1000) % 60)
//   const minutes = Math.floor((t / 1000 / 60) % 60)
//   const hours = Math.floor((t / (1000 * 60 * 60)) % 24)
//   const days = Math.floor(t / (1000 * 60 * 60 * 24))

//   if (t <= 0) {
//     return <Notes>Contest has ended.</Notes>
//   }

//   return `${days} day${days !== 1 && 's'} ${hours} hour${
//     hours !== 1 && 's'
//   } ${minutes} minute${minutes !== 1 && 's'} ${seconds} second${
//     seconds !== 1 && 's'
//   } remaining`
// }

// const Notes = styled.p`
//   padding: 0 0 20px 0;
// `
