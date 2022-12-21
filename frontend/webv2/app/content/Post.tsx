import ReactMarkdown from 'react-markdown'
import { usePost } from '@app/content/api'

interface Props {
  slug: string
}

export const Post = ({ slug }: Props) => {
  const page = usePost(slug)

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
      <ReactMarkdown children={page.data.content} />
    </>
  )
}
