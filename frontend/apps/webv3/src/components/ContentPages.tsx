import { Link } from '@tanstack/react-router'
import { Button, EmptyState, PageHeader, Pagination, prose } from 'paper-ui'
import ReactMarkdown from 'react-markdown'
import rehypeSanitize from 'rehype-sanitize'
import remarkGfm from 'remark-gfm'
import { blogHref, blogPageSize, type BlogPage, type BlogPost, type CmsPage } from '../lib/content'
import { formatDate } from '../lib/format'
import { renderLink } from './router-link'


/** The publish date as a UTC day, so server and browser render the same text. */
const published = (post: BlogPost) => formatDate(new Date(post.publishedAt).toISOString())

/** A content read that failed or answered something unreadable; retrying reads it again. */
export function ContentUnavailable({ query }: { query: { isFetching: boolean; refetch: () => Promise<unknown> } }) {
  return (
    <EmptyState
      headingLevel={1}
      title="This page couldn’t load"
      description="Check your connection, then try again."
      actions={<Button variant="outline" loading={query.isFetching} onClick={() => void query.refetch()}>Try again</Button>}
    />
  )
}

const article = 'min-w-0 max-w-article'
const postLink = 'text-ink decoration-muted hover:text-link hover:decoration-current'

export function CmsPageArticle({ page }: { page: CmsPage }) {
  return (
    <article className={article}>
      <PageHeader title={page.title} className="mb-8" />
      <div className={prose({ className: 'content-prose' })} dangerouslySetInnerHTML={{ __html: page.html }} />
    </article>
  )
}

export function BlogPostArticle({ post }: { post: BlogPost }) {
  return (
    <article className={article}>
      <PageHeader
        title={post.title}
        className="mb-8"
        classNames={{ lead: '[&_time]:text-metadata' }}
        lead={<time dateTime={post.publishedAt}>{published(post)}</time>}
        backLink={{ href: blogHref.list(), label: 'Blog' }}
        renderLink={renderLink}
      />
      <div className={prose({ className: 'content-prose' })}>
        {/* As in Webv2: Markdown without raw HTML, through rehype-sanitize's default schema. */}
        <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]}>{post.content}</ReactMarkdown>
      </div>
    </article>
  )
}

/** One page of the blog: page 1 leads with the newest post's excerpt, as in Webv2. */
export function BlogList({ blog, page }: { blog: BlogPage; page: number }) {
  const totalPages = Math.ceil(blog.total / blogPageSize)
  const lead = page === 1 ? blog.posts[0] : undefined
  const older = lead ? blog.posts.slice(1) : blog.posts

  return (
    <div className={article}>
      <PageHeader title="Blog" className="mb-8" />
      {blog.total === 0 ? (
        <EmptyState headingLevel={2} title="No blog posts yet" description="Check back soon for updates." />
      ) : blog.posts.length === 0 ? (
        <EmptyState
          headingLevel={2}
          title={`There is no page ${page}`}
          description={`The blog has ${totalPages} ${totalPages === 1 ? 'page' : 'pages'}.`}
          actions={<Link to={blogHref.list()} className="text-link">Newest posts</Link>}
        />
      ) : null}

      {lead ? (
        <section className="blog-lead mb-8 border-b border-rule pb-8" aria-labelledby="blog-lead-title">
          <h2 id="blog-lead-title" className="m-0 font-display text-2xl leading-135 wrap-break-word">
            <Link className={postLink} to={blogHref.post(lead.slug)}>{lead.title}</Link>
          </h2>
          <p className="mb-4 mt-1 text-metadata text-muted"><time dateTime={lead.publishedAt}>{published(lead)}</time></p>
          <p className="mb-5 mt-0 max-w-measure-prose leading-180 wrap-break-word">{excerpt(lead.content, 500)}</p>
        </section>
      ) : null}

      {older.length > 0 ? (
        <section aria-label="Older posts">
          <ul className="blog-posts m-0 mb-6 list-none p-0">
            {older.map(post => (
              <li key={post.slug} className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 py-3.5 [&+&]:border-t [&+&]:border-rule">
                <Link className={`${postLink} min-w-0 font-semibold wrap-anywhere`} to={blogHref.post(post.slug)}>{post.title}</Link>
                <time className="flex-none text-sm text-muted" dateTime={post.publishedAt}>{published(post)}</time>
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {totalPages > 1 && page <= totalPages ? (
        <Pagination totalPages={totalPages} currentPage={page} getHref={blogHref.list} renderLink={renderLink} label="Blog pages" />
      ) : null}
    </div>
  )
}

/** Webv2's plain-text excerpt of a post's Markdown, cut at a word boundary. */
function excerpt(markdown: string, maxLength: number): string {
  const text = markdown
    .replace(/#+\s/g, '')
    .replace(/\*\*(.+?)\*\*/g, '$1')
    .replace(/\*(.+?)\*/g, '$1')
    .replace(/\[(.+?)\]\(.+?\)/g, '$1')
    .replace(/!\[.*?\]\(.+?\)/g, '')
    .replace(/`{1,3}[^`]*`{1,3}/g, '')
    .replace(/>\s/g, '')
    .replace(/[-*]\s/g, '')
    .replace(/\n+/g, ' ')
    .trim()

  if (text.length <= maxLength) return text
  return text.slice(0, maxLength).replace(/\s\S*$/, '') + '…'
}
