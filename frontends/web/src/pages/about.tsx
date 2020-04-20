import React from 'react'
import Head from 'next/head'
import PostOrPageDetail from '../app/blog/pages/PostOrPageDetail'
import { ContentContainer } from '../app/ui/components'

const About = () => (
  <>
    <Head>
      <title>Tadoku - About</title>
    </Head>
    <ContentContainer>
      <PostOrPageDetail slug="about" />
    </ContentContainer>
  </>
)

export default About
