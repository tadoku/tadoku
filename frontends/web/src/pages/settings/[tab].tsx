import React from 'react'
import Head from 'next/head'
import { SettingsTab } from '../../app/user/interfaces'
import Settings from '../../app/user/pages/Settings'
import { useRouter } from 'next/router'
import { ContentContainer } from '../../app/ui/components'

const SettingsPage = () => {
  const router = useRouter()
  const currentTab = router.query.tab as SettingsTab

  return (
    <>
      <Head>
        <title>Tadoku - Settings</title>
      </Head>
      <ContentContainer>
        <Settings tab={currentTab} />
      </ContentContainer>
    </>
  )
}

export default SettingsPage
