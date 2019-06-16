import React from 'react'
import { SettingsTab } from '../interfaces'
import styled from 'styled-components'
import Constants from '../../ui/Constants'
import Link from 'next/link'
import { Button } from '../../ui/components'
import { IconProp } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

interface Props {
  activeTab: SettingsTab
}

const SettingsSidebar = ({ activeTab }: Props) => {
  return (
    <Container>
      <Heading>Account settings</Heading>
      <SettingsList>
        {/* <SettingsItem active={activeTab === SettingsTab.ChangePassword}>
          <Link
            as={`/settings/change-password`}
            href={`/settings?tab=change-password`}
          >
            <Button plain icon="cog" disabled>
              Settings
            </Button>
          </Link>
        </SettingsItem> */}
        <SettingsLink
          activeTab={activeTab}
          tab={SettingsTab.ChangePassword}
          icon="cog"
          label="Change Password"
        />
      </SettingsList>
    </Container>
  )
}

export default SettingsSidebar

const SettingsLink = ({
  activeTab,
  tab,
  icon,
  label,
}: {
  activeTab: SettingsTab
  tab: SettingsTab
  icon: IconProp
  label: string
}) => (
  <SettingsItem active={activeTab === tab}>
    <Link as={`/settings/${tab}`} href={`/settings?tab=${tab}`}>
      <StyledLink href="">
        <Icon icon={icon} />
        {label}
      </StyledLink>
    </Link>
  </SettingsItem>
)

const Icon = styled(FontAwesomeIcon)`
  margin-right: 15px;
  height: 75%;
  width: 75%;
`

const StyledLink = styled.a`
  padding: 4px 12px;
  font-size: 1.1em;
  font-weight: 600;
  height: 48px;
  border-radius: 3px;
  box-sizing: border-box;
  margin: 0 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
`

const Container = styled.div`
  margin-right: 20px;
  padding-right: 20px;
  border-right: 1px solid ${Constants.colors.lightGray};
  max-width: 250px;
`

const Heading = styled.h2``

const SettingsList = styled.ul`
  list-style: none;
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  width: 100%;
`

const SettingsItem = styled.li`
  margin: 0;
  padding: 0;

  button,
  a {
    width: 100%;
    padding: 0 10px;
    justify-content: flex-start;

    &:disabled {
      color: inherit;
      opacity: 1;
    }
  }

  ${({ active }: { active: boolean }) =>
    active &&
    `
    border-left: 2px solid ${Constants.colors.primary};
`}
`
