import React from 'react'
import Head from 'next/head'
import styled from 'styled-components'
import PostOrPageDetail from '../app/blog/pages/PostOrPageDetail'

const About = () => (
  <>
    <Head>
      <title>Tadoku - About</title>
    </Head>
    <Container>
      <PostOrPageDetail slug="about" />
    </Container>
  </>
)

const Container = styled.div`
  margin: 0 auto;
  max-width: 700px;
`

export default About
