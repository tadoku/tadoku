import React from 'react'
import BlogApi from '../api'
import { useCachedApiState } from '../../cache'
import { PostOrPage } from '../interfaces'
import BlogPost from '../components/BlogPost'
import { postOrPageCollectionSerializer } from '../transform'

// TODO: Replace this with proper api when we have one
const LatestBlogPost = () => {
  const { data: posts } = useCachedApiState<PostOrPage[]>({
    cacheKey: `blog_list?i=2`,
    defaultValue: [],
    fetchData: BlogApi.posts.list,
    dependencies: [],
    serializer: postOrPageCollectionSerializer,
  })

  if (!posts || !posts[0]) {
    return null
  }

  const latestPost = posts[0]

  return <BlogPost key={latestPost.slug} post={latestPost} />
}

export default LatestBlogPost
