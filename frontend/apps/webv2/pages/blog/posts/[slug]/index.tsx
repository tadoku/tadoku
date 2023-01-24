import type { NextPage } from 'next'
import { Breadcrumb } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PostDetail } from '@app/content/Post'
import { useRouter } from 'next/router'
import { usePost } from '@app/content/api'
import { routes } from '@app/common/routes'
import Head from 'next/head'

interface Props {}

const BlogPost: NextPage<Props> = () => {
  const router = useRouter()
  const { slug } = router.query

  const post = usePost(slug as string)

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
      <Head>
        <title>Blog - {post.data.title} - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'Blog', href: routes.blogList() },
            { label: post.data.title, href: routes.blogPost(post.data.slug) },
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
