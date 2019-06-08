import React from 'react'
import { User } from '../../user/interfaces'

interface Props {
  user: User
}

export const UserInNavigation = ({ user }: Props) => (
  <span>Welcome, {user.displayName}</span>
)
