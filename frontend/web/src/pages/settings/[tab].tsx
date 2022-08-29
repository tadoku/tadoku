import React from 'react'
import Head from 'next/head'
import { SettingsTab } from '@app/user/interfaces'
import Settings from '@app/user/pages/Settings'
import { useRouter } from 'next/router'
import { PageContainer } from '@app/ui/components/Layout'

const SettingsPage = () => {
  const router = useRouter()
  const currentTab = router.query.tab as SettingsTab

  return (
    <>
      <Head>
        <title>Tadoku - Settings</title>
      </Head>
      <PageContainer>
        <Settings tab={currentTab} />
      </PageContainer>
    </>
  )
}

export default SettingsPage
