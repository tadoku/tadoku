/** Keeps a page out of search results; for private and action pages. The canonical link and `og:url` come from the root route. */
export const noindex = { name: 'robots', content: 'noindex' }

/**
 * The document title and the description, Open Graph and Twitter card tags that search results and link previews show.
 * `name` is the page's own name, which the title follows with ` · Tadoku`.
 */
export function sharingMeta(name: string, description: string, title = `${name} · Tadoku`) {
  return [
    { title },
    { name: 'description', content: description },
    { property: 'og:type', content: 'website' },
    { property: 'og:title', content: name },
    { property: 'og:description', content: description },
    { name: 'twitter:card', content: 'summary' },
    { name: 'twitter:title', content: name },
    { name: 'twitter:description', content: description },
  ]
}

const maxDescription = 160
const entities: Record<string, string> = { amp: '&', lt: '<', gt: '>', quot: '"', apos: "'", nbsp: ' ' }

/** Text of at most 160 characters, cut at a word with an ellipsis, or `fallback` when there is none. */
function excerpt(text: string, fallback: string) {
  const plain = text.replace(/\s+/g, ' ').trim()
  if (!plain) return fallback
  if (plain.length <= maxDescription) return plain
  const cut = plain.slice(0, maxDescription - 1)
  return `${cut.slice(0, cut.lastIndexOf(' ') > 0 ? cut.lastIndexOf(' ') : cut.length)}…`
}

/** The character a reference such as `&amp;`, `&#39;` or `&#x27;` stands for, or the reference itself when unknown. */
function decodeReference(reference: string, code: string) {
  if (!code.startsWith('#')) return entities[code.toLowerCase()] ?? reference
  const hex = code[1] === 'x' || code[1] === 'X'
  const point = Number.parseInt(code.slice(hex ? 2 : 1), hex ? 16 : 10)
  return point > 0 && point <= 0x10ffff ? String.fromCodePoint(point) : reference
}

/** A description from sanitized CMS HTML: its text, with tags dropped and character references decoded. */
export function htmlDescription(html: string, fallback: string) {
  return excerpt(html.replace(/<[^>]*>/g, ' ').replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, decodeReference), fallback)
}

/** A description from Markdown: its text, without images, link targets, HTML or block and emphasis markers. */
export function markdownDescription(markdown: string, fallback: string) {
  const text = markdown
    .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/<[^>]*>/g, ' ')
    .replace(/^\s{0,3}(#{1,6}|>|[-*+]|\d+\.)\s+/gm, '')
    .replace(/[*_`~]/g, '')
  return excerpt(text, fallback)
}

/** A contest's own description, or its title and dates when it has none. */
export function contestDescription(contest: { title: string; description?: string; contest_start: string; contest_end: string }) {
  const fallback = `${contest.title}, a Tadoku contest from ${contest.contest_start} to ${contest.contest_end}: standings, activity and participants.`
  return markdownDescription(contest.description ?? '', fallback)
}
