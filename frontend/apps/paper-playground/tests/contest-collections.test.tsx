import { beforeEach, describe, expect, it } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { initialContests, type SampleContest } from '../src/data'
import { PlaygroundProvider, type Viewer } from '../src/state'
import { ContestsPage } from '../src/pages/ContestsPage'

beforeEach(() => localStorage.clear())

function open(path: string, viewer: Viewer, contests: SampleContest[] = initialContests, joinedContests = ['round5', 'reading-circle', 'friends']) {
  localStorage.setItem('tadoku-paper-playground-v1', JSON.stringify({ viewer, contests, joinedContests, logs: [] }))
  return render(<MemoryRouter initialEntries={[path]}><PlaygroundProvider><Routes>
    {['official', 'community', 'participating', 'managed', 'mine'].map(collection => <Route key={collection} path={`/contests/${collection}`} element={<ContestsPage />} />)}
  </Routes></PlaygroundProvider></MemoryRouter>)
}
const collections = () => within(screen.getByRole('navigation', { name: 'Contest collections' }))

describe('separate participating and managed contest collections', () => {
  it('lists only joined contests, including unlisted contests, on the participating page', async () => {
    const user = userEvent.setup()
    open('/contests/participating', 'participant')
    expect(collections().getByRole('link', { name: 'Participating' })).toHaveAttribute('aria-current', 'page')
    for (const title of ['2026 Round 5', 'September reading circle', 'Our Sunday reading group']) expect(screen.getByRole('link', { name: title })).toBeVisible()
    for (const title of ['Japanese novel club', 'Autumn in French', '2026 Round 6']) expect(screen.queryByRole('link', { name: title })).not.toBeInTheDocument()
    await user.type(screen.getByRole('searchbox', { name: 'Find a contest' }), 'Sunday')
    expect(screen.getByRole('status')).toHaveTextContent('1 contest match your filters')
    expect(screen.queryByRole('link', { name: '2026 Round 5' })).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Clear filters' }))
    expect(screen.getByRole('link', { name: '2026 Round 5' })).toBeVisible()
  })

  it('separates owned and moderated contests from participation, and preserves the mine address', async () => {
    const user = userEvent.setup()
    const contests = initialContests.map(contest => contest.id === 'novels' ? { ...contest, moderatorIds: ['anton'] } : contest)
    open('/contests/mine', 'participant', contests)
    expect(collections().getByRole('link', { name: 'Managed by me' })).toHaveAttribute('aria-current', 'page')
    for (const title of ['Japanese novel club', 'Autumn in French', 'Our Sunday reading group']) expect(screen.getByRole('link', { name: title })).toBeVisible()
    expect(screen.queryByRole('link', { name: '2026 Round 5' })).not.toBeInTheDocument()
    await user.click(collections().getByRole('link', { name: 'Participating' }))
    expect(screen.getByRole('link', { name: '2026 Round 5' })).toBeVisible()
    expect(screen.queryByRole('link', { name: 'Autumn in French' })).not.toBeInTheDocument()
  })

  it('hides the managed tab for organizers without assigned contests while keeping creation available', () => {
    open('/contests/community', 'organizer', initialContests.filter(contest => contest.ownerId !== 'anton'))
    expect(collections().queryByRole('link', { name: 'Managed by me' })).not.toBeInTheDocument()
    expect(collections().getByRole('link', { name: 'Participating' })).toBeVisible()
    expect(screen.getByRole('link', { name: 'Create contest' })).toBeVisible()
  })

  it('offers browsing when the signed-in account has no registrations', () => {
    open('/contests/participating', 'participant', initialContests, [])
    expect(screen.getByRole('heading', { name: 'You haven’t joined a contest yet.' })).toBeVisible()
    expect(screen.getByRole('link', { name: 'Explore official contests' })).toBeVisible()
    expect(screen.queryByRole('button', { name: 'Clear filters' })).not.toBeInTheDocument()
  })

  it('hides personal collections from guests and explains direct participating links', () => {
    open('/contests/participating', 'guest')
    expect(collections().queryByRole('link', { name: 'Participating' })).not.toBeInTheDocument()
    expect(collections().queryByRole('link', { name: 'Managed by me' })).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Sign in' })).toHaveAttribute('href', '/sign-in?next=%2Fcontests%2Fparticipating')
    expect(screen.queryByRole('link', { name: 'Our Sunday reading group' })).not.toBeInTheDocument()
  })

  it('keeps an explicitly chosen collection when selecting the organizer sample', () => {
    open('/contests/participating?scenario=creator', 'organizer')
    expect(collections().getByRole('link', { name: 'Participating' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('link', { name: '2026 Round 5' })).toBeVisible()
  })
})
