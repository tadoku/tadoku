import React, { FunctionComponent } from 'react'
import Link from 'next/link'
import { format } from 'date-fns'
import styled from 'styled-components'

import { Contest } from '@app/contest/interfaces'
import {
  Table,
  TableHeading,
  TableHeadingCell,
  Row,
  RowAnchor,
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

interface Props {
  contests: Contest[]
}

const Manage = () => {
  const { data: contests, status: statusContests } = useCachedApiState<
    Contest[]
  >({
    cacheKey: `contests_admin?i=1`,
    defaultValue: [],
    fetchData: () => {
      return ContestApi.getAll()
    },
    dependencies: [],
    serializer: contestCollectionSerializer,
  })

  if (!isReady([statusContests])) {
    return <p>Loading...</p>
  }

  return (
    <>
      <HeaderContainer>
        <PageTitle>Manage contests</PageTitle>
        <ActionContainer></ActionContainer>
      </HeaderContainer>
      <ContestList contests={contests} />
    </>
  )
}

export default Manage

const ContestList = ({ contests }: Props) => {
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
          <>
            <SubHeading>{year}</SubHeading>
            <ContestListGroup contests={grouped[year]} />
          </>
        ))}
    </>
  )
}

const ContestListGroup = ({ contests }: Props) => {
  return (
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
          <Row key={contest.id}>
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
                <Button onClick={() => {}} icon="edit" plain>
                  Edit
                </Button>
              </ActionButtonContainer>
            </Cell>
          </Row>
        ))}
      </tbody>
    </Table>
  )
}

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
  margin: 0;
  font-size: 15px;

  button {
    margin: 0 20px;
  }
`
