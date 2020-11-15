import React, { FunctionComponent } from 'react'
import styled from 'styled-components'
import Link from 'next/link'

import { Contest } from '../interfaces'
import {
  Table,
  TableHeading,
  TableHeadingCell,
  ClickableRow,
  RowAnchor,
  Cell,
} from '@app/ui/components/Table'
import { format } from 'date-fns'

interface Props {
  contests: Contest[]
}

const ContestList = ({ contests }: Props) => {
  return (
    <Table>
      <thead>
        <TableHeading>
          <TableHeadingCell>Round</TableHeadingCell>
          <TableHeadingCell>Starting date</TableHeadingCell>
          <TableHeadingCell>Ending date</TableHeadingCell>
        </TableHeading>
      </thead>
      <tbody>
        {contests.map(contest => (
          <ClickableRow key={contest.id}>
            <Cell>
              <RowLink contest={contest}>{contest.description}</RowLink>
            </Cell>
            <Cell>
              <RowLink contest={contest}>
                {format(contest.start, 'MMMM do')}
              </RowLink>
            </Cell>
            <Cell>
              <RowLink contest={contest}>
                {format(contest.end, 'MMMM do')}
              </RowLink>
            </Cell>
          </ClickableRow>
        ))}
      </tbody>
    </Table>
  )
}

export default ContestList

const RowLink: FunctionComponent<{ contest: Contest }> = ({
  contest,
  children,
}) => (
  <Link href={`/contests/${contest.id}/ranking`} passHref>
    <RowAnchor href="">{children}</RowAnchor>
  </Link>
)
