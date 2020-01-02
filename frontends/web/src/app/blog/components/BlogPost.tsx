import React from 'react'
import styled from 'styled-components'
import { PostOrPage } from '../interfaces'
import { PageTitle } from '../../ui/components'

const BlogPost = ({ post }: { post: PostOrPage }) => (
  <Container>
    <PageTitle>{post.title}</PageTitle>
    <Content dangerouslySetInnerHTML={{ __html: post.html }} />
  </Container>
)

export default BlogPost

const Container = styled.div`
  max-width: 700px;
  margin: 0 auto;
`

const Content = styled.div``
