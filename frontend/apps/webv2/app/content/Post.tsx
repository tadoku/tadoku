import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Post } from '@app/content/api'

interface Props {
  post: Post
}

export const PostDetail = ({ post }: Props) => (
  <div className={`auto-format`}>
    <h1 className="title mt-0 mb-4">{post.title}</h1>
    <PostBody post={post} />
  </div>
)

export const PostBody = ({ post }: Props) => (
  <ReactMarkdown children={post.content} remarkPlugins={[remarkGfm]} />
)
