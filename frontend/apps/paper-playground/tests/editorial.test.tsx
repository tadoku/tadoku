import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it } from 'vitest'
import { GuidePage } from '../src/pages/GuidePage'
import { HomePage } from '../src/pages/HomePage'
import { initialLogs } from '../src/data'
import { contestEnd, contestStart, formatDateRange, formatDateTime } from '../src/dates'
import { PlaygroundProvider, usePlayground } from '../src/state'

beforeEach(() => localStorage.clear())

function renderHome(scenario: string) {
  return render(<MemoryRouter initialEntries={[`/?scenario=${scenario}`]}><PlaygroundProvider><HomePage /></PlaygroundProvider></MemoryRouter>)
}

describe('homepage journeys', () => {
  it('offers contest registration to a guest and links each standing to the correct profile', () => {
    renderHome('guest-live')
    expect(screen.getByRole('link', { name: 'Join Round 5' })).toHaveAttribute('href', '/contests/round5/registration')
    expect(screen.getByRole('link', { name: 'Kai' })).toHaveAttribute('href', '/users/kai?contest=round5')
    expect(screen.getByRole('link', { name: 'Full leaderboard' })).toHaveAttribute('href', '/contests/round5/leaderboard')
  })

  it('shows a participant outside the top five with their true rank and a contextual progress link', () => {
    renderHome('participant-live')
    const standing = screen.getByLabelText('Your standing in Round 5')
    expect(standing).toHaveTextContent('8')
    expect(standing).toHaveTextContent('600')
    expect(within(standing).getByRole('link', { name: 'Anton You' })).toHaveAttribute('href', '/users/anton?contest=round5')
    expect(screen.getByRole('link', { name: 'Log activity' })).toHaveAttribute('href', '/logs/new')
  })

  it('separates the upcoming round dates from its registration deadline', () => {
    renderHome('upcoming')
    expect(screen.getByRole('link', { name: 'Join Round 6' })).toHaveAttribute('href', '/contests/round6/registration')
    expect(within(screen.getByRole('region', { name: '2026 Round 6' })).getByText(formatDateRange(contestStart('2026-11-01'), contestEnd('2026-11-14')), { normalizer: text => text })).toBeInTheDocument()
    expect(screen.getByText(`Registration closes ${formatDateTime(contestEnd('2026-11-07'))}.`, { normalizer: text => text })).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('uses the September recap between rounds and preserves its ended context', () => {
    renderHome('between')
    expect(screen.getByRole('heading', { name: '2026 Round 5 recap' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'See final board' })).toHaveAttribute('href', '/contests/round5/leaderboard?scenario=ended')
    expect(screen.getByRole('link', { name: 'Log activity' })).toBeInTheDocument()
  })

  it('recovers the unavailable board without a page refresh', async () => {
    const user = userEvent.setup()
    renderHome('unavailable')
    expect(screen.getByRole('heading', { name: 'The board could not be loaded' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Retry standings' }))
    expect(screen.getByRole('table', { name: 'Current Round 5 top five' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'The board could not be loaded' })).not.toBeInTheDocument()
  })

  it('keeps the signed-in member when the registration action changes routes', async () => {
    function RegistrationRole() { const { viewer } = usePlayground(); return <output aria-label="Registration viewer">{viewer}</output> }
    const user = userEvent.setup()
    render(<MemoryRouter initialEntries={['/?scenario=member-open']}><PlaygroundProvider><Routes><Route path="/" element={<HomePage />} /><Route path="/contests/round5/registration" element={<RegistrationRole />} /></Routes></PlaygroundProvider></MemoryRouter>)
    await user.click(screen.getByRole('link', { name: 'Join Round 5' }))
    expect(screen.getByLabelText('Registration viewer')).toHaveTextContent('member')
  })

  it('updates the home standing and rank when an activity is submitted or deleted', async () => {
    function LogActions() {
      const { saveLog, removeLog } = usePlayground()
      const log = { ...initialLogs[0], id: 'new-reading', amount: 500, submissions: [{ contestId: 'round5', score: 500, basis: '500 pages × 1 point = 500 points.' }] }
      return <><button onClick={() => saveLog(log)}>Submit reading</button><button onClick={() => removeLog(log.id)}>Delete reading</button></>
    }
    const user = userEvent.setup()
    render(<MemoryRouter initialEntries={['/?scenario=participant-live']}><PlaygroundProvider><LogActions /><HomePage /></PlaygroundProvider></MemoryRouter>)
    await user.click(screen.getByRole('button', { name: 'Submit reading' }))
    expect(screen.getByLabelText('Your standing in Round 5')).toHaveTextContent('1,100')
    expect(screen.getByLabelText('Your standing in Round 5')).toHaveTextContent('3')
    await user.click(screen.getByRole('button', { name: 'Delete reading' }))
    expect(screen.getByLabelText('Your standing in Round 5')).toHaveTextContent('600')
    expect(screen.getByLabelText('Your standing in Round 5')).toHaveTextContent('8')
  })
})

describe('practical guide', () => {
  it('searches answer text as well as questions and clears back to a focused search', async () => {
    const user = userEvent.setup()
    render(<MemoryRouter><PlaygroundProvider><GuidePage page="faq" /></PlaygroundProvider></MemoryRouter>)
    const search = screen.getByRole('searchbox', { name: 'Search these questions' })
    expect(screen.getByRole('status')).toHaveTextContent('6 questions')
    await user.type(search, 'unlisted')
    expect(screen.getByRole('status')).toHaveTextContent('1 question')
    expect(screen.getByText('Does a private contest hide everything from others?')).toBeInTheDocument()
    await user.clear(search)
    await user.type(search, 'zebras')
    expect(screen.getByRole('heading', { name: 'No questions match that search.' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Clear search' }))
    expect(search).toHaveFocus()
    expect(search).toHaveValue('')
    expect(screen.getByRole('status')).toHaveTextContent('6 questions')
  })

  it('uses keyboard-operable native answer disclosures with real contextual routes', async () => {
    const user = userEvent.setup()
    render(<MemoryRouter><PlaygroundProvider><GuidePage page="faq" /></PlaygroundProvider></MemoryRouter>)
    const summary = screen.getByText('Why does my log have more than one score?')
    await user.click(summary)
    expect(summary.closest('details')).toHaveAttribute('open')
    expect(screen.getByRole('link', { name: 'Inspect a sample score breakdown' })).toHaveAttribute('href', '/logs/reading-konbini')
  })

  it('changes all three navigation structures while preserving the article', async () => {
    function IAControls() {
      const { setIa } = usePlayground()
      return <><button onClick={() => setIa('separate')}>Four pages</button><button onClick={() => setIa('tasks')}>Task help</button><button onClick={() => setIa('guide')}>Guide navigation</button></>
    }
    const user = userEvent.setup()
    render(<MemoryRouter><PlaygroundProvider><IAControls /><GuidePage page="manual" /></PlaygroundProvider></MemoryRouter>)
    expect(within(screen.getByRole('navigation', { name: 'Guide' })).getByRole('link', { name: 'Start here' })).toHaveAttribute('aria-current', 'page')
    await user.click(screen.getByRole('button', { name: 'Four pages' }))
    expect(within(screen.getByRole('navigation', { name: 'Learn' })).getByRole('link', { name: 'Manual' })).toHaveAttribute('aria-current', 'page')
    await user.click(screen.getByRole('button', { name: 'Task help' }))
    expect(within(screen.getByRole('navigation', { name: 'Help' })).getByRole('link', { name: 'Join & track' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('heading', { level: 1, name: 'Make your first entry.' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Guide navigation' }))
    expect(screen.getByRole('navigation', { name: 'Guide' })).toBeInTheDocument()
  })
})
