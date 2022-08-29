import React from 'react'
import Head from 'next/head'
import BlogsList from '@app/blog/pages/BlogsList'
import { PageContainer } from '@app/ui/components/Layout'

const Blog = () => (
  <>
    <Head>
      <title>Tadoku - Blog</title>
    </Head>

    <PageContainer>
      <BlogsList />
    </PageContainer>
  </>
)

export default Blog
