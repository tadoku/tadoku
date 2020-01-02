import React from 'react'
import Head from 'next/head'
import styled from 'styled-components'
import PostOrPageDetail from '../app/blog/pages/PostOrPageDetail'

const Manual = () => (
  <>
    <Head>
      <title>Tadoku - Manual</title>
    </Head>
    <Container>
      <PostOrPageDetail slug="manual" />
    </Container>
  </>
)

const Container = styled.div`
  margin: 0 auto;
  max-width: 700px;
`

export default Manual
