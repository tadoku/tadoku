import React from 'react'
import Head from 'next/head'
import { ExpressNextContext } from '../app/interfaces'
import { SettingsTab } from '../app/user/interfaces'
import Settings from '../app/user/pages/Settings'

interface Props {
  tab: SettingsTab
}

const SettingsPage = ({ tab }: Props) => {
  return (
    <>
      <Head>
        <title>Tadoku - Settings</title>
      </Head>
      <Settings tab={tab} />
    </>
  )
}

SettingsPage.getInitialProps = async ({ req, query }: ExpressNextContext) => {
  const defaultTab = SettingsTab.Profile
  if (req && req.params) {
    const { tab } = req.params
    return { tab: tab || defaultTab }
  }

  if (query.tab) {
    const { tab } = query
    return { tab: tab || defaultTab }
  }

  return { tab: defaultTab }
}

export default SettingsPage
