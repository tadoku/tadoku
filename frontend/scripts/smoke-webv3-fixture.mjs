// The synthetic, GET-only Tadoku API that scripts/smoke-webv3-image.mjs runs the image against.
import { createServer } from 'node:http'

export const contest = {
  id: '11111111-1111-4111-8111-111111111111', title: 'Container fixture contest',
  contest_start: '2020-01-01', contest_end: '2099-12-31', registration_end: '2099-12-31',
  private: false, official: true,
  allowed_languages: process.env.WEBV3_LANGUAGE_FIXTURE === 'array' ? [{ code: 'jpn', name: 'Japanese' }] : null,
  language_code_allow_list: process.env.WEBV3_LANGUAGE_FIXTURE === 'array' ? ['jpn'] : null,
  allowed_activities: [], activity_type_id_allow_list: [],
}
const learner = { id: '01000000-0000-0000-0000-000000000000', display_name: 'Fixture learner', created_at: '2024-03-01T08:00:00Z' }
// The learner's profile year: one Japanese reading day in whichever year is requested.
function profileYear(kind, year) {
  const score = 12.5
  switch (kind) {
    case 'activity': return { total_updates: 3, scores: [{ date: `${year}-01-02`, score }] }
    case 'scores': return { overall_score: score, scores: [{ language_code: 'jpn', language_name: 'Japanese', score }] }
    case 'activity-split': return { activities: [{ activity_id: 1, activity_name: 'Reading', score }] }
    case 'contest-registrations': return {
      registrations: [{ id: '22222222-2222-4222-8222-222222222222', contest_id: contest.id, user_id: learner.id,
        user_display_name: learner.display_name, languages: [], contest }],
      total_size: 1, next_page_token: '',
    }
  }
}
// The learner's logs, newest first: 51 fill two pages of 50; the third is tagged NSFW, which hides its description.
const learnerLogs = Array.from({ length: 51 }, (_, index) => ({
  id: `33333333-3333-4333-8333-${String(index).padStart(12, '0')}`, user_id: learner.id,
  user_display_name: learner.display_name, description: `Fixture reading ${51 - index}`,
  language: { code: 'jpn', name: 'Japanese' }, activity: { id: 1, name: 'Reading' }, unit_name: 'Page',
  tags: index === 2 ? ['nsfw'] : [], amount: 4, modifier: 1, score: 4,
  created_at: new Date(Date.parse('2024-03-01T08:00:00Z') + (51 - index) * 3_600_000).toISOString(), deleted: false,
}))
// Twelve published posts, newest first, so the blog has two pages; the CMS page holds markup DOMPurify must remove.
const posts = Array.from({ length: 12 }, (_, index) => ({
  id: `00000000-0000-4000-8000-${String(index + 1).padStart(12, '0')}`, slug: `fixture-post-${12 - index}`,
  title: `Fixture post ${12 - index}`, content: `## Fixture\n\nPost **${12 - index}** from the container fixture.`,
  published_at: `2026-01-${String(24 - index).padStart(2, '0')}T08:00:00Z`,
}))
const cmsPage = {
  id: '00000000-0000-4000-8000-000000000401', slug: 'dev-welcome', title: 'Fixture welcome',
  html: '<p>Fixture page body.</p><img src="data:," alt="" onerror="window.__xss=1"><script>window.__xss=2</script>',
}
const announcement = {
  id: '00000000-0000-4000-8000-000000000501', namespace: 'tadoku', title: 'Fixture notice',
  content: 'The **container fixture** announcement.', style: 'info', href: '/pages/dev-welcome',
  starts_at: '2020-01-01T00:00:00Z', ends_at: '2099-12-31T00:00:00Z',
}
export const writes = []
/**
 * `delayMs` holds every answer that long, so a check can keep a page request open; `status` answers every read with
 * that status instead of data. `requestIds` collects each read's `X-Request-Id`.
 */
export const fixtureOptions = { delayMs: 0, status: 0 }
export const requestIds = []
export const fixture = createServer((request, response) => setTimeout(() => answer(request, response), fixtureOptions.delayMs))
function answer(request, response) {
  requestIds.push(request.headers['x-request-id'])
  if (fixtureOptions.status) {
    response.writeHead(fixtureOptions.status).end()
    return
  }
  if (request.method !== 'GET') {
    writes.push(`${request.method} ${request.url}`)
    response.writeHead(405).end()
    return
  }
  const url = new URL(request.url, 'http://fixture')
  const path = url.pathname
  let data
  if (path.endsWith('/latest-official') || path === `/immersion/contests/${contest.id}`) data = contest
  else if (path.endsWith('/summary')) data = { participant_count: 1, language_count: 1, total_score: 13319.6455 }
  // Imported accounts keep their source IDs without a UUID version check, so standings can hold a version-0 user ID.
  else if (path.endsWith('/leaderboard') || /^\/immersion\/leaderboard\/(global|yearly\/\d+)$/.test(path)) data = {
    // One participant: later pages are empty.
    entries: Number(url.searchParams.get('page') ?? 0) > 0 ? [] : [{ rank: 1, user_id: '01000000-0000-0000-0000-000000000000', user_display_name: 'Fixture learner', score: 13319.6455, is_tie: false }],
    total_size: 1, next_page_token: '',
  }
  // The fixture contest is official, so community lists are empty.
  else if (path === `/content/pages/tadoku/${cmsPage.slug}`) data = cmsPage
  else if (path === '/content/announcements/tadoku/active') data = { announcements: [announcement] }
  else if (path === '/content/posts/tadoku') {
    const size = Number(url.searchParams.get('page_size') ?? 10)
    const page = Number(url.searchParams.get('page') ?? 0)
    data = { posts: posts.slice(page * size, (page + 1) * size), total_size: posts.length, next_page_token: '' }
  }
  else if (path.startsWith('/content/posts/tadoku/')) data = posts.find(post => path.endsWith(`/${post.slug}`))
  else if (path === '/immersion/contests') data = url.searchParams.get('official') === 'false'
    ? { contests: [], total_size: 0, next_page_token: '' }
    : { contests: [contest], total_size: 1, next_page_token: '' }
  else if (path === `/immersion/users/${learner.id}/profile`) data = learner
  else if (path === `/immersion/users/${learner.id}/logs`) {
    const size = Number(url.searchParams.get('page_size') ?? 50)
    const start = Number(url.searchParams.get('page') ?? 0) * size
    data = { logs: learnerLogs.slice(start, start + size), total_size: learnerLogs.length, next_page_token: '' }
  }
  else if (path.startsWith('/immersion/logs/')) data = learnerLogs.find(log => path === `/immersion/logs/${log.id}`)
  // The contest's updates are the learner's logs, two pages of 50; the learner is its only participant.
  else if (path === `/immersion/contests/${contest.id}/logs`) {
    const size = Number(url.searchParams.get('page_size') ?? 50)
    const start = Number(url.searchParams.get('page') ?? 0) * size
    const logs = [learner.id, null].includes(url.searchParams.get('user_id')) ? learnerLogs : []
    data = { logs: logs.slice(start, start + size), total_size: logs.length, next_page_token: '' }
  }
  else if (path === `/immersion/contests/${contest.id}/profile/${learner.id}/scores`) data = {
    overall_score: 13319.6455, scores: [{ language_code: 'jpn', language_name: 'Japanese', score: 13319.6455 }],
    registration: { id: '22222222-2222-4222-8222-222222222222', contest_id: contest.id, user_id: learner.id,
      user_display_name: learner.display_name, languages: [{ code: 'jpn', name: 'Japanese' }], contest },
  }
  else if (path === `/immersion/contests/${contest.id}/profile/${learner.id}/activity`) data = {
    rows: [{ date: '2024-03-03', language_code: 'jpn', score: 13319.6455 }],
  }
  else {
    const match = path.match(new RegExp(`^/immersion/users/${learner.id}/([a-z-]+)/(\\d{4})$`))
    if (match) data = profileYear(match[1], match[2])
  }
  if (!data) { response.writeHead(404).end(); return }
  response.writeHead(200, { 'Content-Type': 'application/json' }).end(JSON.stringify(data))
}
await new Promise(resolve => fixture.listen(0, '0.0.0.0', resolve))
export const apiPort = fixture.address().port
