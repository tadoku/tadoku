import * as stylex from '@stylexjs/stylex'
import { Link } from '@tanstack/react-router'
import { Button, EmptyState, PageHeader, Pagination, Prose, ProseContent, linkStyles, proseElements } from 'paper-ui'
import { border, color, font, fontSize, fontWeight, lineHeight, measure, space, type } from 'paper-ui/tokens.stylex'
import ReactMarkdown from 'react-markdown'
import rehypeSanitize from 'rehype-sanitize'
import remarkGfm from 'remark-gfm'
import { blogHref, blogPageSize, type BlogPage, type BlogPost, type CmsPage } from '../lib/content'
import { formatDate } from '../lib/format'
import { renderLink } from './router-link'
import { contentStyles } from './page-styles'


const styles = stylex.create({
  date: { fontSize: type.metadataSize },
  lead: {
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: 'solid',
    borderBlockEndWidth: border.static,
    marginBlockEnd: space[32],
    paddingBlockEnd: space[32],
  },
  leadTitle: { margin: 0, fontFamily: font.display, fontSize: fontSize[24], lineHeight: lineHeight.heading, overflowWrap: 'break-word', },
  postLink: { color: { default: color.textInk, ':hover': color.textLink }, textDecorationColor: { default: color.textMuted, ':hover': 'currentColor' } },
  leadDate: { marginInline: 0, color: color.textMuted, fontSize: type.metadataSize, marginBlockEnd: space[16], marginBlockStart: space[4], },
  excerpt: { marginInline: 0, lineHeight: lineHeight.prose, marginBlockEnd: space[20], marginBlockStart: 0, maxInlineSize: measure.prose, overflowWrap: 'break-word', },
  posts: { padding: 0, listStyle: 'none', marginInline: 0, marginBlockEnd: space[24], marginBlockStart: 0, },
  post: {
    paddingBlock: space[14],
    paddingInline: 0,
    alignItems: 'baseline',
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: { default: border.static, ':first-child': 0 },
    columnGap: space[16],
    display: 'flex',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    rowGap: space[4],
  },
  postTitle: { fontWeight: fontWeight.semibold, minInlineSize: 0, overflowWrap: 'anywhere' },
  postDate: { flex: 'none', color: color.textMuted, fontSize: fontSize[14], },
})


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

export function CmsPageArticle({ page }: { page: CmsPage }) {
  return (
    <article {...stylex.props(contentStyles.article)}>
      <PageHeader title={page.title} />
      <Prose><ProseContent nodes={page.body} /></Prose>
    </article>
  )
}

export function BlogPostArticle({ post }: { post: BlogPost }) {
  return (
    <article {...stylex.props(contentStyles.article)}>
      <PageHeader
        title={post.title}
        lead={<time dateTime={post.publishedAt} {...stylex.props(styles.date)}>{published(post)}</time>}
        backLink={{ href: blogHref.list(), label: 'Blog' }}
        renderLink={renderLink}
      />
      <Prose>
        {/* As in Webv2: Markdown without raw HTML, through rehype-sanitize's default schema. */}
        <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]} components={proseElements}>{post.content}</ReactMarkdown>
      </Prose>
    </article>
  )
}

/** One page of the blog: page 1 leads with the newest post's excerpt, as in Webv2. */
export function BlogList({ blog, page }: { blog: BlogPage; page: number }) {
  const totalPages = Math.ceil(blog.total / blogPageSize)
  const lead = page === 1 ? blog.posts[0] : undefined
  const older = lead ? blog.posts.slice(1) : blog.posts

  return (
    <div {...stylex.props(contentStyles.article)}>
      <PageHeader title="Blog" />
      {blog.total === 0 ? (
        <EmptyState headingLevel={2} title="No blog posts yet" description="Check back soon for updates." />
      ) : blog.posts.length === 0 ? (
        <EmptyState
          headingLevel={2}
          title={`There is no page ${page}`}
          description={`The blog has ${totalPages} ${totalPages === 1 ? 'page' : 'pages'}.`}
          actions={<Link to={blogHref.list()} {...stylex.props(linkStyles.base)}>Newest posts</Link>}
        />
      ) : null}

      {lead ? (
        <section {...stylex.props(styles.lead)} aria-labelledby="blog-lead-title" data-blog-lead="">
          <h2 id="blog-lead-title" {...stylex.props(styles.leadTitle)}>
            <Link to={blogHref.post(lead.slug)} {...stylex.props(linkStyles.base, styles.postLink)}>{lead.title}</Link>
          </h2>
          <p {...stylex.props(styles.leadDate)}><time dateTime={lead.publishedAt}>{published(lead)}</time></p>
          <p {...stylex.props(styles.excerpt)}>{excerpt(lead.content, 500)}</p>
        </section>
      ) : null}

      {older.length > 0 ? (
        <section aria-label="Older posts">
          <ul {...stylex.props(styles.posts)} data-blog-posts="">
            {older.map(post => (
              <li key={post.slug} {...stylex.props(styles.post)}>
                <Link to={blogHref.post(post.slug)} {...stylex.props(linkStyles.base, styles.postLink, styles.postTitle)}>{post.title}</Link>
                <time dateTime={post.publishedAt} {...stylex.props(styles.postDate)}>{published(post)}</time>
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
