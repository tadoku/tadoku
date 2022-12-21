import type { NextPage } from 'next'
import Breadcrumb from 'tadoku-ui/components/Breadcrumb'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PostDetail } from '@app/content/Post'
import { useRouter } from 'next/router'
import { usePost } from '@app/content/api'

interface Props {}

const BlogPost: NextPage<Props> = () => {
  const router = useRouter<'/blog/post/[slug]'>()
  const { slug } = router.query

  const post = usePost(slug)

  if (post.isLoading || post.isIdle) {
    return <p>Loading...</p>
  }

  if (post.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: '/', IconComponent: HomeIcon },
            { label: 'Blog', href: '/blog' },
            { label: post.data.title, href: `/blog/${post.data.slug}` },
          ]}
        />
      </div>
      <div>
        <PostDetail post={post.data} />
      </div>
    </>
  )
}

export default BlogPost
