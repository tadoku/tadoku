import React from 'react'
import Head from 'next/head'
import PostOrPageDetail from '@app/blog/pages/PostOrPageDetail'
import { PageContainer } from '@app/ui/components/Layout'

const Manual = () => (
  <>
    <Head>
      <title>Tadoku - Manual</title>
    </Head>
    <PageContainer>
      <PostOrPageDetail slug="manual" />
    </PageContainer>
  </>
)

export default Manual
