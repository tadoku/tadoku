import React, { useState } from 'react'
import { format } from 'date-fns'
import styled from 'styled-components'
import media from 'styled-media-query'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

import { Ranking, RankingRegistration } from '../interfaces'
import Leaderboard from '../components/Leaderboard'
import RankingApi from '../api'
import { Contest } from '@app/contest/interfaces'
import { Button, PageTitle, SubHeading } from '@app/ui/components'
import { User } from '@app/session/interfaces'
import JoinContestModal from '../components/modals/JoinContestModal'
import { useCachedApiState, ApiFetchStatus } from '../../cache'
import { rankingCollectionSerializer } from '../transform/ranking'
import { isRegisteredForContest, canJoinContest } from '../domain'
import SubmitPagesButton from '../components/SubmitPagesButton'
import Constants from '@app/ui/Constants'

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
        <ContestPeriod contest={contest} />
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

  ${media.lessThan('medium')`
    flex-direction: column;
    margin-bottom: 20px;

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

const Description = styled(SubHeading)`
  margin-top: 10px;
  margin-bottom: 10px;
`
const ContestPeriod = ({ contest }: { contest: Contest }) => {
  return (
    <ContestPeriodContainer>
      <ContestPeriodDate>
        <ContestPeriodLabel>Starting</ContestPeriodLabel>
        <ContestPeriodValue>
          {format(contest.start, 'MMMM dd')}
        </ContestPeriodValue>
      </ContestPeriodDate>
      <Icon icon="arrow-right" />
      <ContestPeriodDate>
        <ContestPeriodLabel>Ending</ContestPeriodLabel>
        <ContestPeriodValue>
          {format(contest.end, 'MMMM dd')}
        </ContestPeriodValue>
      </ContestPeriodDate>
    </ContestPeriodContainer>
  )
}

const ContestPeriodContainer = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  box-sizing: border-box;
  background: ${Constants.colors.light};

  ${media.lessThan('small')`width: 100%; margin-top: 5px;`}
`

const Icon = styled(FontAwesomeIcon)`
  color: ${Constants.colors.nonFocusText};
  opacity: 0.4;
  margin: 0 20px;
`

const ContestPeriodDate = styled.div``

const ContestPeriodLabel = styled.div`
  font-size: 12px;
  text-transform: uppercase;
  font-weight: bold;
  color: ${Constants.colors.nonFocusText};
`
const ContestPeriodValue = styled.div`
  font-weight: bold;
  font-size: 13px;
`
