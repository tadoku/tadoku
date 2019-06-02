import React from 'react'
import { RankingRegistration } from '../../../ranking/interfaces'
import { SubmitPages } from '../../../ranking/navigation/SubmitPages'
import SignOut from '../../../session/navigation/LogOut'

export const ActiveUserNavigationBar = ({
  registration,
}: {
  registration: RankingRegistration | undefined
}) => (
  <>
    <SubmitPages registration={registration} />
    <SignOut />
  </>
)
