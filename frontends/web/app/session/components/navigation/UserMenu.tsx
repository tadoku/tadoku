import React from 'react'
import { User } from '../../interfaces'
import LogOutLink from './LogOut'
import Dropdown, { DropdownItem } from '../../../ui/components/Dropdown'

interface Props {
  user: User
}

const UserProfile = ({ user }: Props) => (
  <Dropdown label={user.displayName}>
    <DropdownItem>
      <LogOutLink></LogOutLink>
    </DropdownItem>
  </Dropdown>
)

export default UserProfile
