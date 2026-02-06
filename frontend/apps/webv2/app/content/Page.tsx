import { routes } from '@app/common/routes'
import { usePage } from '@app/content/api'
import DOMPurify from 'dompurify'
import { HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb, Loading } from 'ui'

interface Props {
  slug: string
}

export const Page = ({ slug }: Props) => {
  const page = usePage(slug)

  if (page.isLoading || page.isIdle) {
    return <Loading />
  }

  if (page.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  return (
    <>
      <Head>
        <title>{page.data.title} - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: page.data.title, href: routes.blogPage(slug) },
          ]}
        />
      </div>

      <div className="max-w-3xl">
        <h1 className="title my-4">{page.data.title}</h1>
        <div className="auto-format">
          <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(page.data.html) }} />
        </div>
      </div>
    </>
  )
}
