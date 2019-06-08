import React from 'react'
import { User } from '../../../user/interfaces'

interface Props {
  user: User
}

const UserProfile = ({ user }: Props) => (
  <span>Welcome, {user.displayName}</span>
)

export default UserProfile
