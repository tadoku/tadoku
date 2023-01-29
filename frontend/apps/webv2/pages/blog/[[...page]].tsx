import { usePostList } from '@app/content/api'
import { PostDetail } from '@app/content/Post'
import { useRouter } from 'next/router'
import { Breadcrumb, Loading, Pagination } from 'ui'
import { HomeIcon } from '@heroicons/react/24/outline'
import { useState } from 'react'
import Link from 'next/link'
import { getQueryStringIntParameter } from '@app/common/router'
import { routes } from '@app/common/routes'
import Head from 'next/head'

// TODO: Rewrite page and navigation logic
const BlogIndex = () => {
  const router = useRouter()

  const [page, setPage] = useState(() => {
    return getQueryStringIntParameter(router.query.page, 1)
  })

  const pageSize = 10
  const list = usePostList({ pageSize, page: page - 1 })

  const navigateToPage = async (page: number) => {
    setPage(page)
    await router.push(routes.blogList(page))
  }

  if (list.isLoading || list.isIdle) {
    return <Loading />
  }

  if (list.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const totalPages = Math.ceil(list.data.total_size / pageSize)

  return (
    <>
      <Head>
        <title>Blog - page {page} - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'Blog', href: routes.blogList() },
          ]}
        />
      </div>
      <div className="space-y-8">
        <div className="grid grid-cols-1 gap-8 lg:grid-cols-2">
          {list.data.posts.map(p => (
            <div className="card max-h-96	overflow-hidden relative" key={p.id}>
              <Link
                href={{
                  pathname: '/blog/posts/[slug]',
                  query: { slug: p.slug },
                }}
                className="absolute left-0 right-0 top-80 h-16 flex justify-center items-center font-bold bg-gradient-to-t from-white via-white to-transparent"
              >
                Read more...
              </Link>
              <PostDetail post={p} key={p.id} />
            </div>
          ))}
        </div>
        {totalPages > 1 ? (
          <Pagination
            currentPage={page}
            totalPages={totalPages}
            onClick={navigateToPage}
          />
        ) : null}
      </div>
    </>
  )
}

export default BlogIndex
