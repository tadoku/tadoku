import React, { useState } from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'

import { ContestLog, Ranking, RankingRegistration } from '../interfaces'
import Leaderboard from '../components/Leaderboard'
import ContestUpdates from '../components/ContestUpdates'
import RankingApi from '../api'
import { Contest } from '@app/contest/interfaces'
import { Button, PageTitle, SubHeading } from '@app/ui/components'
import { User } from '@app/session/interfaces'
import JoinContestModal from '../components/modals/JoinContestModal'
import { useCachedApiState, ApiFetchStatus } from '../../cache'
import { rankingCollectionSerializer } from '../transform/ranking'
import { isRegisteredForContest, canJoinContest } from '../domain'
import SubmitPagesButton from '../components/SubmitPagesButton'
import ContestPeriod from '../components/ContestPeriod'
import { contestLogCollectionSerializer } from '../transform/contest-log'

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

  const { data: rankings, status: statusRanking } = useCachedApiState<
    Ranking[]
  >({
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

  const { data: logs, status: statusLogs } = useCachedApiState<ContestLog[]>({
    cacheKey: `recent_contest_logs?i=1&contest_id=${contest.id}`,
    defaultValue: [],
    fetchData: () => {
      if (!contest) {
        return new Promise<ContestLog[]>(resolve => resolve([]))
      }

      return RankingApi.getRecentLogs(contest.id)
    },
    dependencies: [contest?.id, effectCount],
    serializer: contestLogCollectionSerializer,
  })

  const isRegistered = isRegisteredForContest(registration, contest)
  const canJoin = canJoinContest(user, registration, contest)

  return (
    <>
      <Header>
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
      </Header>

      <TwoColumn>
        <Leaderboard
          rankings={rankings}
          loading={statusRanking === ApiFetchStatus.Loading}
        />
        <Sidebar>
          <ContestPeriod contest={contest} />
          <ContestUpdates
            logs={logs}
            loading={statusLogs == ApiFetchStatus.Loading}
          />
        </Sidebar>
      </TwoColumn>
    </>
  )
}

export default RankingOverview

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  margin-bottom: 15px;

  h1 {
    margin: 0;
  }

  ${media.lessThan('medium')`
    flex-direction: column;

    > button {
      margin: 10px 0;
    }
  `}

  ${media.lessThan('small')`
    margin-bottom: 20px;

    > button {
      width: 100%;
      box-sizing: border-box;
      margin: 10px 0;
    }
  `}
`

const TwoColumn = styled.div`
  display: flex;
  flex-direction: row;
  margin-bottom: 30px;
  width: 100%;

  ${media.lessThan('medium')`
    flex-direction: column;
    margin-bottom: 20px;
  `}
`

const Sidebar = styled.div`
  display: flex;
  flex-direction: column;
  margin-left: 30px;
  flex-basis: 260px;
  flex-grow: 0;
  flex-shrink: 0;
  box-sizing: border-box;

  ${media.lessThan('medium')`
    flex-basis: inherit;
    margin: 0 0 30px 0;
  `}
`

const Description = styled(SubHeading)`
  margin-top: 10px;
  margin-bottom: 10px;
`
