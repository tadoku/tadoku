import React from 'react'
import { User } from '../../interfaces'
import LogOutLink from './LogOut'
import Dropdown, { DropdownItem } from '../../../ui/components/Dropdown'
import { Button } from '../../../ui/components'
import Link from 'next/link'
import { SettingsTab } from '../../../user/interfaces'

interface Props {
  user: User
}

const UserProfile = ({ user }: Props) => (
  <Dropdown label={user.displayName}>
    <DropdownItem>
      <Link href={`/settings/${SettingsTab.Profile}`}>
        {/* TODO: Remove span once https://github.com/zeit/next.js/issues/7915 is fixed */}
        <span>
          <Button plain icon="cog">
            Settings
          </Button>
        </span>
      </Link>
    </DropdownItem>
    <DropdownItem>
      <LogOutLink />
    </DropdownItem>
  </Dropdown>
)

export default UserProfile
