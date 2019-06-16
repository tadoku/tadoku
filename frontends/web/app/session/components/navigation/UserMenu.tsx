import React, { useState } from 'react'
import { User } from '../../interfaces'
import { Button } from '../../../ui/components'
import SignOutLink from './LogOut'
import styled from 'styled-components'
import Constants from '../../../ui/Constants'

interface Props {
  user: User
}

const UserProfile = ({ user }: Props) => {
  const [isMenuOpen, setIsMenuOpen] = useState(true)

  return (
    <Container>
      <Button
        onClick={() => setIsMenuOpen(!isMenuOpen)}
        icon={isMenuOpen ? 'chevron-up' : 'chevron-down'}
        plain
        alignIconRight
      >
        {user.displayName}
      </Button>
      {isMenuOpen && (
        <DropDown>
          <DropDownItem>
            <SignOutLink />
          </DropDownItem>
        </DropDown>
      )}
    </Container>
  )
}

export default UserProfile

const Container = styled.div`
  position: relative;
`

const DropDown = styled.ul`
  position: absolute;
  top: 40px;
  width: 100%;

  list-style: none;
  box-sizing: border-box;
  margin: 0;
  padding: 0;

  background: ${Constants.colors.light};
  box-shadow: 0 2px 4px 0px rgba(0, 0, 0, 0.08);
  border-radius: 2px;
  border: 1px solid ${Constants.colors.lightGray};
`

const DropDownItem = styled.li`
  padding: 2px 12px;
  margin: 0;

  &:hover {
    background: ${Constants.colors.lightGray};
  }

  & + & {
    border-top: 1px solid ${Constants.colors.lightGray};
  }
`
