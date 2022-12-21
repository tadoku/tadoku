import ReactMarkdown from 'react-markdown'
import { Post } from '@app/content/api'
import { PostDetail } from './Post'

interface Props {
  posts: Post[]
}

export const PostList = ({ posts }: Props) =>
  posts.map(p => <PostDetail post={p} key={p.id} />)
