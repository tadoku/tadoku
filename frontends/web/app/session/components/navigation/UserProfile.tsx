import React, { useState } from 'react'
import { User } from '../../interfaces'
import { Button } from '../../../ui/components'
import SignOutLink from './LogOut'

interface Props {
  user: User
}

const UserProfile = ({ user }: Props) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  return (
    <>
      <Button
        onClick={() => setIsMenuOpen(!isMenuOpen)}
        icon="chevron-down"
        plain
        alignIconRight
      >
        {user.displayName}
      </Button>
      {isMenuOpen && (
        <ul>
          <li>
            <SignOutLink />
          </li>
        </ul>
      )}
    </>
  )
}

export default UserProfile
