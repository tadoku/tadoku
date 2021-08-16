import React from 'react'
import Head from 'next/head'
import BlogsList from '@app/blog/pages/BlogsList'

const Blog = () => (
  <>
    <Head>
      <title>Tadoku - Blog</title>
    </Head>
    <BlogsList />
  </>
)

export default Blog
