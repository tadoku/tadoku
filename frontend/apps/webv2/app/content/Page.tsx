import { usePage } from '@app/content/api'

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
      <h1 className="title my-4">{page.data.title}</h1>
      <div dangerouslySetInnerHTML={{ __html: page.data.html }} />
    </>
  )
}
