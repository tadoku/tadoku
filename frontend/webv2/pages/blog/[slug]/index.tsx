import type { NextPage } from 'next'
import Breadcrumb from 'tadoku-ui/components/Breadcrumb'
import { HomeIcon } from '@heroicons/react/20/solid'
import { Page } from '@app/content/Page'
import { Post } from '@app/content/Post'
import { useRouter } from 'next/router'

interface Props {}

const BlogPost: NextPage<Props> = () => {
  const router = useRouter()
  const { slug } = router.query

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
        <Post slug={slug!.toString()} />
      </div>
    </>
  )
}

export default BlogPost
