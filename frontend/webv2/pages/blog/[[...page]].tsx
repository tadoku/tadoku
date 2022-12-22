import { usePostList } from '@app/content/api'
import { PostDetail } from '@app/content/Post'
import { useRouter } from 'next/router'
import Breadcrumb from 'tadoku-ui/components/Breadcrumb'
import Pagination from 'tadoku-ui/components/Pagination'
import { HomeIcon } from '@heroicons/react/24/outline'
import { useState } from 'react'
import Link from 'next/link'

const BlogIndex = () => {
  const router = useRouter<'/blog/[[...page]]'>()

  const [page, setPage] = useState(() => {
    if (!router.query.page) {
      return 1
    }

    const idx = parseInt(router.query.page.toString())

    if (isNaN(idx)) {
      return 1
    }

    return idx
  })

  const pageSize = 10
  const list = usePostList(pageSize, page - 1)

  const navigateToPage = async (page: number) => {
    setPage(page)
    await router.push({
      pathname: `/blog/[[...page]]`,
      query: { page: page.toString() },
    })
  }

  if (list.isLoading || list.isIdle) {
    return <p>Loading...</p>
  }

  if (list.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const totalPages = Math.ceil(list.data.totalSize / pageSize)

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: '/', IconComponent: HomeIcon },
            { label: 'Blog', href: '/blog' },
          ]}
        />
      </div>
      <div className="space-y-8">
        <div className="grid grid-cols-2 gap-8">
          {list.data.posts.map(p => (
            <div className="card max-h-96	overflow-hidden relative">
              <Link
                href={{
                  pathname: '/blog/post/[slug]',
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
