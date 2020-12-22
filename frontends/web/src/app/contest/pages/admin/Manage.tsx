import React, { Fragment, useState } from 'react'
import Link from 'next/link'
import { format } from 'date-fns'
import styled from 'styled-components'

import { Contest } from '@app/contest/interfaces'
import {
  Table,
  TableHeading,
  TableHeadingCell,
  Row,
  Cell,
} from '@app/ui/components/Table'
import {
  Button,
  ButtonContainer,
  ButtonLink,
  PageTitle,
  SubHeading,
} from '@app/ui/components'
import { isReady, useCachedApiState } from '@app/cache'
import ContestApi from '@app/contest/api'
import { contestCollectionSerializer } from '@app/contest/transform'
import Constants from '@app/ui/Constants'
import EditContestFormModal from '@app/contest/components/modals/EditContestFormModal'
import { isContestEditable } from '@app/ranking/domain'
import NewContestFormModal from '@app/contest/components/modals/NewContestFormModal'

interface Props {
  contests: Contest[]
  editContest: (contest: Contest) => void
}

const Manage = () => {
  const [effectCount, setEffectCount] = useState(0)
  const { data: contests, status: statusContests } = useCachedApiState<
    Contest[]
  >({
    cacheKey: `contests_admin?i=1`,
    defaultValue: [],
    fetchData: () => {
      return ContestApi.getAll()
    },
    dependencies: [effectCount],
    serializer: contestCollectionSerializer,
  })
  const [selectedContest, setSelectedContest] = useState(
    undefined as Contest | undefined,
  )
  const [createContestModalOpen, setCreateContestModalOpen] = useState(false)

  if (!isReady([statusContests])) {
    return <p>Loading...</p>
  }

  return (
    <>
      <HeaderContainer>
        <PageTitle>Manage contests</PageTitle>
        <ActionContainer>
          <NewContestFormModal
            isOpen={createContestModalOpen}
            onCancel={() => setCreateContestModalOpen(false)}
            onSuccess={() => {
              setCreateContestModalOpen(false)
              setEffectCount(effectCount + 1)
            }}
          />
          <Button onClick={() => setCreateContestModalOpen(true)} primary>
            Create contest
          </Button>
        </ActionContainer>
      </HeaderContainer>
      <EditContestFormModal
        setContest={setSelectedContest}
        contest={selectedContest}
        onCancel={() => setSelectedContest(undefined)}
        onSuccess={() => {
          setSelectedContest(undefined)
          setEffectCount(effectCount + 1)
        }}
      />
      <ContestList contests={contests} editContest={setSelectedContest} />
    </>
  )
}

export default Manage

const ContestList = ({ contests, editContest }: Props) => {
  const grouped = contests.reduce((grouped, contest) => {
    const year = contest.start.getUTCFullYear()
    grouped[year] = grouped[year] || []
    grouped[year].push(contest)
    return grouped
  }, ({} as any) as { [key: string]: Contest[] })

  return (
    <>
      {Object.keys(grouped)
        .sort((a, b) => Number(b) - Number(a))
        .map(year => (
          <Fragment key={year}>
            <SubHeading>{year}</SubHeading>
            <ContestListGroup
              contests={grouped[year]}
              editContest={editContest}
            />
          </Fragment>
        ))}
    </>
  )
}

const ContestListGroup = ({ contests, editContest }: Props) => (
  <Table>
    <thead>
      <TableHeading>
        <TableHeadingCell>Round</TableHeadingCell>
        <TableHeadingCell>Starting date</TableHeadingCell>
        <TableHeadingCell>Ending date</TableHeadingCell>
        <TableHeadingCell>Actions</TableHeadingCell>
      </TableHeading>
    </thead>
    <tbody>
      {contests.map(contest => (
        <Row key={contest.id} fontSize="16px">
          <Cell>{contest.description}</Cell>
          <Cell>{format(contest.start, 'MMMM do')}</Cell>
          <Cell>{format(contest.end, 'MMMM do')}</Cell>
          <Cell style={{ width: '1px', whiteSpace: 'nowrap', padding: 0 }}>
            <ActionButtonContainer>
              <Link href={`/contests/${contest.id}/ranking`} passHref>
                <ButtonLink icon="eye" plain>
                  View
                </ButtonLink>
              </Link>
              <Button
                onClick={() => editContest(contest)}
                icon="edit"
                plain
                disabled={!isContestEditable(contest)}
              >
                Edit
              </Button>
            </ActionButtonContainer>
          </Cell>
        </Row>
      ))}
    </tbody>
  </Table>
)

const HeaderContainer = styled.div`
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  border-bottom: 2px solid ${Constants.colors.lightGray};
  padding-bottom: 20px;

  h1 {
    margin: 0;
  }
`

const ActionContainer = styled.div`
  display: flex;
  margin-right: -5px;

  button {
    margin: 0 5px;
  }
`

const ActionButtonContainer = styled(ButtonContainer)`
  margin: 0 0 0 20px;

  button {
    margin: 0 20px;
  }
`
