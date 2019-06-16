import React from 'react'
import { User } from '../../interfaces'
import { Button } from '../../../ui/components'

interface Props {
  user: User
}

const UserProfile = ({ user }: Props) => (
  <Button icon="chevron-down" plain alignIconRight>
    {user.displayName}
  </Button>
)

export default UserProfile
