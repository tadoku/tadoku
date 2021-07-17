import React from 'react'
import styled from 'styled-components'
import { PostOrPage } from '../interfaces'

const BlogPage = ({ post }: { post: PostOrPage }) => (
  <>
    <Title>{post.title}</Title>
    <Content dangerouslySetInnerHTML={{ __html: post.html }} />
  </>
)

export default BlogPage

const Content = styled.div`
  h2 {
    font-size: 1.3em;
    margin-top: 2em;
    margin-bottom: 0.6em;
  }

  h3 {
    font-size: 1.2em;
    margin-top: 1.7em;
  }
`

const Title = styled.h2``
