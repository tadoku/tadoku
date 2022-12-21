import { usePostList } from '@app/content/api'
import { PostDetail } from '@app/content/Post'
import { useRouter } from 'next/router'
import Breadcrumb from 'tadoku-ui/components/Breadcrumb'
import Pagination from 'tadoku-ui/components/Pagination'
import { HomeIcon } from '@heroicons/react/24/outline'

interface Props {}

const BlogIndex = () => {
  const router = useRouter<'/blog/[slug]'>()
  const { slug } = router.query

  const pageSize = 2
  const list = usePostList(pageSize, 0)

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
      <div>
        {list.data.posts.map(p => (
          <PostDetail post={p} key={p.id} />
        ))}
        {totalPages > 1 ? (
          <Pagination currentPage={0} totalPages={totalPages} />
        ) : null}
      </div>
    </>
  )
}

export default BlogIndex
