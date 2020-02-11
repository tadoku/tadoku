import React from 'react'
import LogOutLink from './LogOut'
import Dropdown, { DropdownItem } from '../../../ui/components/Dropdown'
import { Button } from '../../../ui/components'
import Link from 'next/link'
import { SettingsTab } from '../../../user/interfaces'

import { RankingRegistration } from '../../../ranking/interfaces'
import { User } from '../../interfaces'

interface Props {
  user: User
  registration: RankingRegistration | undefined
}

const Settings = () => (
  <DropdownItem>
    <Link
      href="/settings/[tab]"
      as={`/settings/${SettingsTab.Profile}`}
      passHref
    >
      {/* TODO: Remove span once https://github.com/zeit/next.js/issues/7915 is fixed */}
      <span>
        <Button plain icon="cog">
          Settings
        </Button>
      </span>
    </Link>
  </DropdownItem>
)

interface ProfileLinkProps {
  user: User
  registration: RankingRegistration
}
const ContestProfile = ({ user, registration }: ProfileLinkProps) => {
  return (
    <DropdownItem>
      <Link
        href="/contest-profile/[tab]/[tab]"
        as={`/contest-profile/${registration.contestId}/${user.id}`}
      >
        {/* TODO: Remove span once https://github.com/zeit/next.js/issues/7915 is fixed */}
        <span>
          <Button plain icon="user">
            Profile
          </Button>
        </span>
      </Link>
    </DropdownItem>
  )
}

const UserMenu = ({ user, registration }: Props) => {
  return registration ? (
    <Dropdown label={user.displayName}>
      <Settings />
      <ContestProfile user={user} registration={registration}></ContestProfile>
      <DropdownItem>
        <LogOutLink />
      </DropdownItem>
    </Dropdown>
  ) : (
    <Dropdown label={user.displayName}>
      <Settings />
      <DropdownItem>
        <LogOutLink />
      </DropdownItem>
    </Dropdown>
  )
}

export default UserMenu
