import ReactMarkdown from 'react-markdown'
import { Post } from '@app/content/api'

interface Props {
  post: Post
}

export const PostDetail = ({ post }: Props) => (
  <>
    <h1 className="title mt-0 mb-4">{post.title}</h1>
    <ReactMarkdown children={post.content} />
  </>
)
