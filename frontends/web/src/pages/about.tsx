import React from 'react'
import Head from 'next/head'
import PostOrPageDetail from '@app/blog/pages/PostOrPageDetail'

const About = () => (
  <>
    <Head>
      <title>Tadoku - About</title>
    </Head>
    <PostOrPageDetail slug="about" />
  </>
)

export default About
