import React from 'react'
import styled from 'styled-components'
import { PostOrPage } from '../interfaces'
import { format } from 'date-fns'
import Constants from '@app/ui/Constants'
import { SubHeading } from '@app/ui/components'

const BlogPost = ({ post }: { post: PostOrPage }) => (
  <Container>
    <Title>{post.title}</Title>
    <Date>{format(post.publishedAt, 'MMMM yyyy')}</Date>
    <Content dangerouslySetInnerHTML={{ __html: post.html }} />
  </Container>
)

export default BlogPost

const Container = styled.div`
  & + & {
    margin-top: 30px;
    border-top: 2px solid ${Constants.colors.lightGray};
    padding-top: 30px;
  }
`

const Content = styled.div`
  a {
    color: ${Constants.colors.primary};
    text-decoration: underline;

    &:hover {
      opacity: 0.7;
    }
  }
`

const Title = styled.h2`
  margin: 0;
`

const Date = styled(SubHeading).attrs({ as: 'h3' })`
  margin-top: 5px;
  margin-bottom: 30px;
  font-size: 16px;
`
