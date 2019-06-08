import React from 'react'
import { User } from '../../interfaces'

interface Props {
  user: User
}

const UserProfile = ({ user }: Props) => <div>Welcome, {user.displayName}</div>

export default UserProfile
