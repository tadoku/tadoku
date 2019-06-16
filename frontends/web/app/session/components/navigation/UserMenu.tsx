import React, { useState } from 'react'
import { User } from '../../interfaces'
import { Button } from '../../../ui/components'
import SignOutLink from './LogOut'
import styled, { keyframes, css } from 'styled-components'
import Constants from '../../../ui/Constants'

interface Props {
  user: User
}

const UserProfile = ({ user }: Props) => {
  const [isMenuOpen, setIsMenuOpen] = useState(true)

  return (
    <Container>
      <DropDown open={isMenuOpen}>
        <DropDownItem>
          <SignOutLink />
        </DropDownItem>
      </DropDown>
      <Button
        onClick={() => setIsMenuOpen(!isMenuOpen)}
        icon={isMenuOpen ? 'chevron-up' : 'chevron-down'}
        plain
        alignIconRight
        style={{ position: 'relative', zIndex: 2, margin: '0 20px' }}
      >
        {user.displayName}
      </Button>
    </Container>
  )
}

export default UserProfile

const Container = styled.div`
  position: relative;
`

const show = keyframes`
  from {
      opacity: 0;
  }
  to {
      opacity: 1;
  }
`

const DropDown = styled.ul`
  display: none;
  position: absolute;
  top: 0;
  z-index: 1;
  width: 102%;
  list-style: none;
  box-sizing: border-box;
  margin: 0;
  padding: 48px 0 0;

  background: ${Constants.colors.light};
  box-shadow: 0 2px 4px 0px rgba(0, 0, 0, 0.08);
  border-radius: 2px;
  border: 1px solid ${Constants.colors.lightGray};

  ${({ open }: { open: boolean }) =>
    open &&
    css`
      display: block;
      animation: ${show} 0.3s ease;
    `}
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
