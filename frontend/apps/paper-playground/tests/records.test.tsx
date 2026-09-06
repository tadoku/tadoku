import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it } from 'vitest'
import { initialContests, initialLogs, scoreLog } from '../src/data'
import { PlaygroundProvider } from '../src/state'
import { filterActivityLogs, ProfilePage, summarizeActivity } from '../src/pages/ProfilePages'
import { LogPage, submissionEligibility, updateSubmissions } from '../src/pages/LogPage'
import { AdminPage } from '../src/pages/AdminPage'

beforeEach(() => localStorage.clear())

describe('Personal records and contest contributions', () => {
  it('uses each contest’s sample rule when adding submissions', () => {
    const reading = { ...initialLogs.find(log => log.id === 'reading-konbini')!, submissions: [] }
    const listening = { ...initialLogs.find(log => log.id === 'listening-teppei')!, submissions: [] }
    expect(updateSubmissions(reading, ['round5', 'reading-circle'], initialContests, ['round5', 'reading-circle']).map(item => item.score)).toEqual([42, 22.5])
    expect(updateSubmissions(listening, ['round5', 'listening-circle'], initialContests, ['round5', 'listening-circle']).map(item => item.score)).toEqual([17.5, 35])
  })

  it('requires tracked reading time for a contest that scores minutes', () => {
    const reading = { ...initialLogs.find(log => log.id === 'personal-reading')!, submissions: [] }
    expect(submissionEligibility(reading, initialContests.find(contest => contest.id === 'reading-circle')!, ['reading-circle'])).toMatch(/time/i)
  })

  it('honors the languages selected during contest registration', () => {
    const reading = { ...initialLogs.find(log => log.id === 'personal-reading')!, language: 'French', submissions: [] }
    expect(submissionEligibility(reading, initialContests.find(contest => contest.id === 'round5')!, ['round5'], { round5: ['Japanese', 'Spanish'] })).toMatch(/registered.*French|French.*registration/i)
  })

  it('keeps closed contest snapshots while editing other submissions', () => {
    const ended = initialLogs.find(log => log.id === 'ended-reading')!
    expect(updateSubmissions(ended, [], initialContests, ['round5', 'reading-circle'])).toEqual(ended.submissions)
    expect(scoreLog(ended)).toBe(48)
    expect(ended.submissions[0].score).toBe(42)
  })

  it('derives the selected year’s totals, chart and breakdown from the same records', () => {
    const logs = filterActivityLogs(initialLogs.filter(log => log.userId === 'anton'), { period: '2025', languages: [], activities: [], query: '' })
    const summary = summarizeActivity(logs)
    expect(logs.length).toBeGreaterThan(0)
    expect(logs.every(log => log.date.startsWith('2025'))).toBe(true)
    expect(summary.total).toBe(summary.daily.reduce((sum, day) => sum + day.value, 0))
    expect(summary.total).toBe(summary.languages.reduce((sum, [, score]) => sum + score, 0))
    expect(summary.total).toBe(summary.activities.reduce((sum, [, score]) => sum + score, 0))
    expect(summary.days).toBe(summary.daily.length)
  })

  it('combines activity filters without changing the source records', () => {
    const filtered = filterActivityLogs(initialLogs, { period: '2026', languages: ['Japanese'], activities: ['Reading'], query: 'chapter 3' })
    expect(filtered.length).toBeGreaterThan(0)
    expect(filtered.every(log => log.language === 'Japanese' && log.activity === 'Reading' && log.date.startsWith('2026') && log.title.includes('chapter 3'))).toBe(true)
    expect(filterActivityLogs(initialLogs, { period: 'all', languages: [], activities: [], query: 'does not exist anywhere' })).toEqual([])
  })
})

function renderRoute(path: string) {
  return render(<MemoryRouter initialEntries={[path]}><PlaygroundProvider><Routes><Route path="/logs/:logId" element={<LogPage />} /><Route path="/users/:userId" element={<ProfilePage />} /><Route path="/users/:userId/activity" element={<ProfilePage activity />} /><Route path="/admin" element={<AdminPage />} /><Route path="/admin/:section" element={<AdminPage />} /></Routes></PlaygroundProvider></MemoryRouter>)
}

it('keeps another viewer’s submissions and owner controls out of log details', () => {
  renderRoute('/logs/reading-mei?scenario=visitor')
  expect(screen.getByRole('heading', { name: /finished chapter 3/ })).toBeInTheDocument()
  expect(screen.getByText('Contest submissions are visible to the log’s owner.')).toBeInTheDocument()
  expect(screen.queryByRole('link', { name: 'Edit log' })).not.toBeInTheDocument()
  expect(screen.queryByText('More actions')).not.toBeInTheDocument()
  expect(screen.queryByText('No contest submissions')).not.toBeInTheDocument()
  expect(screen.queryByRole('link', { name: '2026 Round 5' })).not.toBeInTheDocument()
})

it('keeps the requested record when no scenario override is present', () => {
  renderRoute('/logs/listening-teppei')
  expect(screen.getByRole('heading', { name: 'Nihongo con Teppei — episode 920' })).toBeInTheDocument()
  expect(screen.getByText('minutes listened')).toBeInTheDocument()
})

it('clears a no-match activity search and its filters', async () => {
  renderRoute('/users/anton/activity?scenario=profile')
  fireEvent.change(screen.getByLabelText('Search activity'), { target: { value: 'no matching book here' } })
  expect(await screen.findByText('No activity matches these filters.')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: 'Clear search and filters' }))
  expect(screen.getByLabelText('Search activity')).toHaveValue('')
  expect(screen.queryByText('No activity matches these filters.')).not.toBeInTheDocument()
  expect(within(screen.getByRole('status')).getByText(/entries/)).toBeInTheDocument()
})

it('keeps deletion cancellable and restores the deleted local sample', async () => {
  const user = userEvent.setup()
  renderRoute('/logs/reading-konbini?scenario=reading')
  await user.click(screen.getByText('More actions'))
  await user.click(screen.getByRole('button', { name: 'Delete log' }))
  const dialog = await screen.findByRole('dialog', { name: 'Delete this log?' })
  await user.click(within(dialog).getByRole('button', { name: 'Keep log' }))
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  expect(screen.getByRole('heading', { name: /finished chapter 3/ })).toBeInTheDocument()
  await user.click(screen.getByRole('button', { name: 'Delete log' }))
  await user.keyboard('{Escape}')
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  await user.click(screen.getByRole('button', { name: 'Delete log' }))
  await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Delete log' }))
  expect(await screen.findByRole('heading', { name: 'Log deleted' })).toHaveFocus()
  await user.click(screen.getByRole('button', { name: 'Restore sample' }))
  expect(screen.getByRole('heading', { name: /finished chapter 3/ })).toBeInTheDocument()
})

it('keeps the selected profile year and contest context in local navigation', async () => {
  const user = userEvent.setup()
  renderRoute('/users/anton?scenario=profile&contest=round5')
  await user.click(screen.getByRole('button', { name: 'Previous year' }))
  expect(screen.getByRole('img', { name: 'Daily activity for 2025' })).toBeInTheDocument()
  expect(screen.getByRole('link', { name: 'Activity log' })).toHaveAttribute('href', '/users/anton/activity?year=2025&contest=round5&scenario=profile')
  expect(screen.getByRole('link', { name: '2026 Round 5 standings' })).toHaveAttribute('href', '/contests/round5/leaderboard')
})

it('saves an admin review and updates the dashboard queue', async () => {
  const user = userEvent.setup()
  renderRoute('/admin')
  await user.click(screen.getAllByRole('button', { name: 'Review' })[0])
  const dialog = await screen.findByRole('dialog', { name: 'Review report' })
  await user.selectOptions(within(dialog).getByLabelText('Status'), 'Resolved')
  await user.click(within(dialog).getByRole('button', { name: 'Save review' }))
  expect(await screen.findByText('Unclear activity description: resolved saved in this workspace.')).toBeInTheDocument()
  expect(screen.getAllByRole('button', { name: 'Review' })).toHaveLength(1)
})
