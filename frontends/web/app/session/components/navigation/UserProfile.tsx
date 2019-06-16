import React, { useState } from 'react'
import { User } from '../../interfaces'
import { Button } from '../../../ui/components'

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
      {isMenuOpen && 'open'}
    </>
  )
}

export default UserProfile
