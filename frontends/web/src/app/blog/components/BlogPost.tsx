import React from 'react'
import styled from 'styled-components'
import { PostOrPage } from '../interfaces'

const BlogPost = ({ post }: { post: PostOrPage }) => (
  <Container>
    <h2>{post.title}</h2>
    <h3>
      {post.publishedAt.getFullYear()} {post.publishedAt.getMonth()}
    </h3>
    <Content dangerouslySetInnerHTML={{ __html: post.html }} />
  </Container>
)

export default BlogPost

const Container = styled.div`
  & + & {
    margin-top: 60px;
  }
`

const Content = styled.div``
