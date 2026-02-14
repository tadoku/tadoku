import ReactMarkdown from 'react-markdown'
import rehypeSanitize from 'rehype-sanitize'
import remarkGfm from 'remark-gfm'
import { DateTime } from 'luxon'
import { Post } from '@app/content/api'

interface Props {
  post: Post
}

export const PostDetail = ({ post }: Props) => (
  <div>
    <h1 className="title mt-0">{post.title}</h1>
    <h2 className="subtitle">{DateTime.fromISO(post.published_at).toLocaleString(DateTime.DATE_FULL)}</h2>
    <div className="auto-format">
      <PostBody post={post} />
    </div>
  </div>
)

export const PostBody = ({ post }: Props) => (
  <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]}>{post.content}</ReactMarkdown>
)
