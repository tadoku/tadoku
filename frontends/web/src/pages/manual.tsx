import React from 'react'
import Head from 'next/head'
import PostOrPageDetail from '../app/blog/pages/PostOrPageDetail'
import { ContentContainer } from '../app/ui/components'

const Manual = () => (
  <>
    <Head>
      <title>Tadoku - Manual</title>
    </Head>
    <ContentContainer>
      <PostOrPageDetail slug="manual" />
    </ContentContainer>
  </>
)

export default Manual
