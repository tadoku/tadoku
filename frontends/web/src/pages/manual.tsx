import React from 'react'
import Head from 'next/head'
import PostOrPageDetail from '@app/blog/pages/PostOrPageDetail'

const Manual = () => (
  <>
    <Head>
      <title>Tadoku - Manual</title>
    </Head>
    <PostOrPageDetail slug="manual" />
  </>
)

export default Manual
