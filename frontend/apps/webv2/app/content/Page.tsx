import { routes } from '@app/common/routes'
import { usePage } from '@app/content/api'
import { HomeIcon } from '@heroicons/react/20/solid'
import { Breadcrumb } from 'ui'

interface Props {
  slug: string
}

export const Page = ({ slug }: Props) => {
  const page = usePage(slug)

  if (page.isLoading || page.isIdle) {
    return <p>Loading...</p>
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
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: page.data.title, href: routes.blogPage(slug) },
          ]}
        />
      </div>
      <h1 className="title my-4">{page.data.title}</h1>
      <div dangerouslySetInnerHTML={{ __html: page.data.html }} />
    </>
  )
}
