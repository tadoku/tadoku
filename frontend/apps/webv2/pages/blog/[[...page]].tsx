import { usePostList } from '@app/content/api'
import { useRouter } from 'next/router'
import { Breadcrumb, Loading, Pagination } from 'ui'
import { HomeIcon } from '@heroicons/react/24/outline'
import { useState } from 'react'
import Link from 'next/link'
import { getQueryStringIntParameter } from '@app/common/router'
import { routes } from '@app/common/routes'
import Head from 'next/head'
import { DateTime } from 'luxon'

const getExcerpt = (markdown: string, maxLength: number): string => {
  const text = markdown
    .replace(/#+\s/g, '')
    .replace(/\*\*(.+?)\*\*/g, '$1')
    .replace(/\*(.+?)\*/g, '$1')
    .replace(/\[(.+?)\]\(.+?\)/g, '$1')
    .replace(/!\[.*?\]\(.+?\)/g, '')
    .replace(/`{1,3}[^`]*`{1,3}/g, '')
    .replace(/>\s/g, '')
    .replace(/[-*]\s/g, '')
    .replace(/\n+/g, ' ')
    .trim()

  if (text.length <= maxLength) return text
  return text.slice(0, maxLength).replace(/\s\S*$/, '') + '…'
}

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
  const posts = list.data.posts
  const heroPost = page === 1 && posts.length > 0 ? posts[0] : undefined
  const olderPosts = heroPost ? posts.slice(1) : posts

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
        {heroPost && (
          <div className="pb-8 border-b border-slate-200">
            <Link
              href={{
                pathname: '/blog/posts/[slug]',
                query: { slug: heroPost.slug },
              }}
            >
              <h1 className="font-serif font-bold text-3xl hover:text-primary transition-colors">
                {heroPost.title}
              </h1>
            </Link>
            <h2 className="subtitle">
              {DateTime.fromISO(heroPost.published_at).toLocaleString(
                DateTime.DATE_FULL,
              )}
            </h2>
            <p className="text-slate-700 text-lg leading-relaxed max-w-3xl">
              {getExcerpt(heroPost.content, 500)}
            </p>
            <div className="mt-5">
              <Link
                href={{
                  pathname: '/blog/posts/[slug]',
                  query: { slug: heroPost.slug },
                }}
                className="btn primary"
              >
                Read more &rarr;
              </Link>
            </div>
          </div>
        )}

        {olderPosts.length > 0 && (
          <div>
          <h2 className="subtitle mb-4">Older posts</h2>
          <ul className="divide-y divide-slate-200">
            {olderPosts.map(p => (
              <li key={p.id}>
                <Link
                  href={{
                    pathname: '/blog/posts/[slug]',
                    query: { slug: p.slug },
                  }}
                  className="flex flex-col gap-1 py-4 sm:flex-row sm:items-baseline sm:justify-between group"
                >
                  <span className="font-serif font-bold text-lg group-hover:text-primary transition-colors">
                    {p.title}
                  </span>
                  <span className="text-sm text-slate-500 sm:ml-4 shrink-0">
                    {DateTime.fromISO(p.published_at).toLocaleString(
                      DateTime.DATE_MED,
                    )}
                  </span>
                </Link>
              </li>
            ))}
          </ul>
          </div>
        )}

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
