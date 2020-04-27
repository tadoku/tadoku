import React from 'react'
import BlogApi from '../api'
import { useCachedApiState } from '../../cache'
import BlogPage from '../components/BlogPage'
import { PostOrPage } from '../interfaces'
import { PostOrPageSerializer } from '../transform'
import { OptionalizeSerializer } from '../../transform'

interface Props {
  slug: string
}

const PostOrPageDetail = ({ slug }: Props) => {
  const { data: page } = useCachedApiState<PostOrPage | undefined>({
    cacheKey: `blog_page?i=3&slug=${slug}`,
    defaultValue: undefined,
    fetchData: () => BlogApi.pages.get(slug),
    dependencies: [],
    serializer: OptionalizeSerializer(PostOrPageSerializer),
  })

  if (!page) {
    return null
  }

  return <BlogPage post={page} />
}

export default PostOrPageDetail
