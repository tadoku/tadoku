import ReactMarkdown, { type Components } from 'react-markdown'
import rehypeSanitize from 'rehype-sanitize'
import remarkGfm from 'remark-gfm'

const components: Components = {
  img: ({ node, ...props }) => {
    void node
    return <img {...props} loading="lazy" decoding="async" />
  },
}

/**
 * CMS Markdown, as Webv2 renders it: GitHub-flavoured, without raw HTML, through rehype-sanitize's default schema.
 * Images load as they near the viewport. The default export lets the shell load it with `lazy()`.
 */
export default function Markdown({ children }: { children: string }) {
  return <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]} components={components}>{children}</ReactMarkdown>
}
