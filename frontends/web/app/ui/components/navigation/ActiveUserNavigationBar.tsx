import React from 'react'
import Link from 'next/link'
import { RankingRegistration } from '../../../ranking/interfaces'
import { SubmitPages } from '../../../ranking/navigation/SubmitPages'
import { NavLink } from './Menu'

export const ActiveUserNavigationBar = ({
  registration,
}: {
  registration: RankingRegistration | undefined
}) => (
  <>
    <SubmitPages registration={registration} />
    <Link href="/sign-out">
      <NavLink href="">Sign out</NavLink>
    </Link>
  </>
)
