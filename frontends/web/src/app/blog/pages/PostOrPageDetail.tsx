import React from 'react'
import BlogApi from '../api'
import { useCachedApiState } from '../../cache'
import BlogPost from '../components/BlogPost'
import { PostOrPage } from '../interfaces'

interface Props {
  slug: string
}

const PostOrPageDetail = ({ slug }: Props) => {
  const { data: page } = useCachedApiState<PostOrPage | undefined>({
    cacheKey: `blog_page?i=1`,
    defaultValue: undefined,
    fetchData: () => BlogApi.pages.get(slug),
    dependencies: [],
  })

  if (!page) {
    return null
  }

  return <BlogPost post={page} />
}

export default PostOrPageDetail
