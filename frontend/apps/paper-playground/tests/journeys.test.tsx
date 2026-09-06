import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ToastProvider } from 'paper-ui'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, expect, it } from 'vitest'
import { LogEditorPage } from '../src/SupportingPages'
import { ContestDetailPage, ContestRegistrationPage } from '../src/pages/ContestsPage'
import { HomePage } from '../src/pages/HomePage'
import { LeaderboardPage } from '../src/pages/LeaderboardPage'
import { LogPage } from '../src/pages/LogPage'
import { PlaygroundProvider } from '../src/state'

beforeEach(() => localStorage.clear())

function open(path: string) {
  render(<MemoryRouter initialEntries={[path]}><PlaygroundProvider><ToastProvider><Routes>
    <Route path="/" element={<HomePage />} />
    <Route path="/contests/:contestId/registration" element={<ContestRegistrationPage />} />
    <Route path="/contests/:contestId" element={<ContestDetailPage />} />
    <Route path="/contests/:contestId/leaderboard" element={<LeaderboardPage />} />
    <Route path="/logs/new" element={<LogEditorPage />} />
    <Route path="/logs/:logId/edit" element={<LogEditorPage />} />
    <Route path="/logs/:logId" element={<LogPage />} />
  </Routes></ToastProvider></PlaygroundProvider></MemoryRouter>)
}

it('lets a participant log their registered German language and filter its contest contribution', async () => {
  const user = userEvent.setup()
  open('/contests/round5/registration?scenario=member')
  await user.click(screen.getByRole('button', { name: 'Remove Spanish' }))
  const languages = screen.getByRole('combobox', { name: /Your languages/ })
  for (const language of ['German', 'Mandarin']) {
    await user.click(languages)
    await user.type(languages, language)
    await user.click(await screen.findByRole('option', { name: language }))
  }
  await user.click(screen.getByRole('button', { name: 'Save registration' }))
  expect(await screen.findByText(/Your languages: Japanese, German, Mandarin/)).toBeVisible()
  await user.click(screen.getByRole('link', { name: 'Log activity' }))
  const language = screen.getByRole('combobox', { name: /Language/ })
  await user.selectOptions(language, 'German')
  expect(within(language).getByRole('option', { name: 'Mandarin' })).toBeInTheDocument()
  await user.type(screen.getByLabelText('What did you read or listen to?'), 'A chapter in German')
  await user.type(screen.getByLabelText(/Pages read/), '10')
  expect(screen.getByRole('checkbox', { name: '2026 Round 5' })).toBeChecked()
  await user.click(screen.getByRole('button', { name: 'Save activity' }))
  expect(await screen.findByRole('heading', { name: 'A chapter in German' })).toBeVisible()
  await user.click(screen.getByRole('link', { name: '2026 Round 5' }))
  await user.click(screen.getByRole('link', { name: 'View standings' }))
  await user.selectOptions(screen.getByRole('combobox', { name: 'Language' }), 'German')
  const standing = within(screen.getByRole('table')).getByRole('row', { name: /Anton/ })
  expect(within(standing).getByRole('cell', { name: '10' })).toBeVisible()
})

it('keeps the between-rounds date through creating, viewing and editing an activity', async () => {
  const user = userEvent.setup()
  open('/?scenario=between')
  expect(screen.getByRole('heading', { name: '2026 Round 5 recap' })).toBeVisible()
  await user.click(screen.getByRole('link', { name: 'Log activity' }))
  expect(screen.getByLabelText(/Activity date/)).toHaveValue('2026-10-01')
  expect(screen.queryByRole('checkbox', { name: '2026 Round 5' })).not.toBeInTheDocument()
  await user.type(screen.getByLabelText('What did you read or listen to?'), 'October reading')
  await user.type(screen.getByLabelText(/Pages read/), '10')
  await user.click(screen.getByRole('button', { name: 'Save activity' }))
  expect(await screen.findByRole('heading', { name: 'October reading' })).toBeVisible()
  expect(screen.getByText(/Viewed as of 1 Oct 2026/)).toBeVisible()
  const edit=screen.getByRole('link', { name: 'Edit log' })
  expect(edit.getAttribute('href')).toContain('asOf=2026-10-01')
  await user.click(edit)
  expect(screen.getByLabelText(/Activity date/)).toHaveValue('2026-10-01')
  expect(screen.getByLabelText(/Activity date/)).toHaveAttribute('max','2026-10-01')
  await user.clear(screen.getByLabelText(/Pages read/))
  await user.type(screen.getByLabelText(/Pages read/), '12')
  await user.click(screen.getByRole('button', { name: 'Save changes' }))
  expect(await screen.findByRole('heading', { name: 'October reading' })).toBeVisible()
  expect(screen.getByText(/Viewed as of 1 Oct 2026/)).toBeVisible()
  expect(screen.getByRole('link', { name: 'Edit log' }).getAttribute('href')).toContain('asOf=2026-10-01')
})
