import React from 'react'
import styled from 'styled-components'
import { Post } from '../domain'

const BlogPost = ({ post }: { post: Post }) => (
  <Container>
    <Title>{post.title}</Title>
    <Content dangerouslySetInnerHTML={{ __html: post.html }} />
  </Container>
)

export default BlogPost

const Container = styled.div``

const Title = styled.h2``
const Content = styled.div``
