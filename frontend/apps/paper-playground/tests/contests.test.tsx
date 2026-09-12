import { beforeEach, describe, expect, it } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { PlaygroundProvider } from '../src/state'
import { LeaderboardPage } from '../src/pages/LeaderboardPage'
import { ContestDetailPage, ContestEditorPage, ContestRegistrationPage, ContestsPage } from '../src/pages/ContestsPage'

beforeEach(() => localStorage.clear())

function open(path: string) {
  return render(<MemoryRouter initialEntries={[path]}><PlaygroundProvider><Routes>
    <Route path="/leaderboard/latest" element={<LeaderboardPage />} />
    <Route path="/leaderboard/yearly/:year" element={<LeaderboardPage />} />
    <Route path="/leaderboard/all-time" element={<LeaderboardPage />} />
    <Route path="/contests/:contestId/leaderboard" element={<LeaderboardPage />} />
    <Route path="/contests/official" element={<ContestsPage />} />
    <Route path="/contests/community" element={<ContestsPage />} />
    <Route path="/contests/mine" element={<ContestsPage />} />
    <Route path="/contests/new" element={<ContestEditorPage />} />
    <Route path="/contests/:contestId/edit" element={<ContestEditorPage />} />
    <Route path="/contests/:contestId/registration" element={<ContestRegistrationPage />} />
    <Route path="/contests/:contestId" element={<ContestDetailPage />} />
  </Routes></PlaygroundProvider></MemoryRouter>)
}

function standing(name: string) {
  return screen.getByRole('table').querySelector<HTMLTableRowElement>(`#standing-${name.toLowerCase()}`)!
}

describe('leaderboard population and context', () => {
  it('starts with filters collapsed, preserves applied filters when closed, and clears them without reopening', async () => {
    const user = userEvent.setup()
    open('/leaderboard/latest?scenario=participant')
    const toggle = screen.getByRole('button', { name: 'Filters' })
    expect(toggle).toHaveAttribute('aria-expanded', 'false')
    expect(document.getElementById(toggle.getAttribute('aria-controls')!)).not.toBeVisible()
    expect(screen.queryByRole('combobox', { name: 'Language' })).not.toBeInTheDocument()
    toggle.focus()
    await user.keyboard('{Enter}')
    expect(toggle).toHaveAttribute('aria-expanded', 'true')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Language' }), 'Japanese')
    await user.type(screen.getByRole('searchbox', { name: 'Find a participant' }), 'Anton')
    expect(standing('Anton').cells[2]).toHaveTextContent('450')
    await user.click(toggle)
    expect(toggle).toHaveAccessibleName('Filters (2)')
    expect(screen.queryByRole('searchbox')).not.toBeInTheDocument()
    expect(screen.getByLabelText('Applied filters')).toHaveTextContent('Japanese · Name: Anton')
    expect(standing('Anton').cells[2]).toHaveTextContent('450')
    await user.click(screen.getByRole('button', { name: 'Clear all' }))
    expect(toggle).toHaveFocus()
    expect(toggle).toHaveAccessibleName('Filters')
    expect(toggle).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByLabelText('Applied filters')).not.toBeInTheDocument()
    expect(standing('Kai').cells[2]).toHaveTextContent('1,400')
    await user.click(toggle)
    expect(screen.getByRole('combobox', { name: 'Language' })).toHaveValue('all')
    expect(screen.getByRole('searchbox', { name: 'Find a participant' })).toHaveValue('')
  })

  it('ranks before name search, preserves ties and zero, and finds the current user beyond page one', async () => {
    const user = userEvent.setup()
    open('/leaderboard/latest?scenario=participant')
    expect(screen.getByRole('table').querySelectorAll('tbody tr')).toHaveLength(6)
    expect(standing('Anton')).toBeNull()
    expect(standing('Elliot').cells[0]).toHaveTextContent('T6')
    await user.click(screen.getByRole('button', { name: 'Find my row' }))
    expect(standing('Anton')).toHaveFocus()
    expect(within(standing('Anton')).getByRole('link', { name: 'Anton' })).toHaveAttribute('href', '/users/anton?contest=round5')
    expect(standing('Anton').cells[0]).toHaveTextContent('8')
    expect(standing('Sora').cells[0]).toHaveTextContent('T6')
    await user.click(screen.getByRole('button', { name: 'Filters' }))
    const search = screen.getByRole('searchbox', { name: 'Find a participant' })
    await user.type(search, 'Anton')
    expect(standing('Anton').cells[0]).toHaveTextContent('8')
    expect(standing('Anton').cells[2]).toHaveTextContent('600')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Language' }), 'Japanese')
    expect(standing('Anton').cells[0]).toHaveTextContent('7')
    expect(standing('Anton').cells[2]).toHaveTextContent('450')
    await user.clear(search)
    await user.type(search, 'Taylor')
    expect(standing('Taylor').cells[2]).toHaveTextContent(/^0$/)
    await user.clear(search)
    await user.type(search, 'not a participant')
    expect(screen.getByText('No participants match that name')).toBeVisible()
    await user.click(screen.getByRole('button', { name: 'Clear search' }))
    expect(search).toHaveFocus()
    expect(screen.getByRole('table').querySelectorAll('tbody tr')).toHaveLength(6)
  })

  it('updates yearly datasets and keeps official aggregate scopes distinct from a selected community contest', async () => {
    const user = userEvent.setup()
    open('/leaderboard/yearly/2026?scenario=participant')
    expect(standing('Kai').cells[2]).toHaveTextContent('11,200')
    expect(within(standing('Kai')).getByRole('link', { name: 'Kai' })).toHaveAttribute('href', '/users/kai?year=2026')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Year' }), '2025')
    expect(screen.getByRole('heading', { name: 'Official standings' })).toBeVisible()
    expect(screen.getByRole('combobox', { name: 'Year' })).toHaveValue('2025')
    expect(standing('Kai').cells[2]).toHaveTextContent('16,800')
    expect(within(standing('Kai')).getByRole('link', { name: 'Kai' })).toHaveAttribute('href', '/users/kai?year=2025')
    await user.click(screen.getByRole('link', { name: 'All time' }))
    expect(standing('Kai').cells[2]).toHaveTextContent('85,400')
    expect(within(standing('Kai')).getByRole('link', { name: 'Kai' })).toHaveAttribute('href', '/users/kai')
  })

  it('does not reuse official fixture scores for an unrelated community contest', () => {
    open('/contests/novels/leaderboard?scenario=participant')
    expect(screen.getByRole('heading', { name: 'Japanese novel club' })).toBeVisible()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
    expect(screen.getByText('The first page is yours to write')).toBeVisible()
    expect(screen.queryByText('Kai')).not.toBeInTheDocument()
  })

  it('keeps registration closure separate from participation and recovers a failed board', async () => {
    const user = userEvent.setup()
    const view = open('/leaderboard/latest?scenario=closed-participant')
    expect(screen.getByRole('link', { name: 'Log activity' })).toBeVisible()
    expect(screen.getByText('Closed to new participants')).toBeVisible()
    expect(screen.queryByRole('link', { name: 'Join this round' })).not.toBeInTheDocument()
    view.unmount()
    open('/leaderboard/latest?scenario=error')
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Find my row' })).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Try again' }))
    expect(screen.getByRole('table')).toBeVisible()
  })
})

describe('contest discovery, joining and organizing', () => {
  it('keeps an organizer’s directly opened collection instead of redirecting it to My contests', () => {
    const view = open('/contests/mine?scenario=creator')
    view.unmount()
    open('/contests/community')
    expect(screen.getByRole('link', { name: 'Japanese novel club' })).toBeVisible()
    expect(screen.queryByRole('link', { name: 'Our Sunday reading group' })).not.toBeInTheDocument()
  })

  it('searches the complete collection, distinguishes live from open registration, and keeps unlisted contests out of discovery', async () => {
    const user = userEvent.setup()
    open('/contests/community?scenario=member')
    expect(screen.queryByText('Our Sunday reading group')).not.toBeInTheDocument()
    await user.type(screen.getByRole('searchbox', { name: 'Find a contest' }), 'listening hour')
    const row = screen.getByRole('link', { name: 'September listening hour' }).closest('article')!
    expect(within(row).getByText('Live now')).toBeVisible()
    expect(within(row).getAllByRole('link', { name: 'View standings' })[0]).toBeVisible()
    expect(within(row).queryByRole('link', { name: 'Join contest' })).not.toBeInTheDocument()
    await user.selectOptions(screen.getByRole('combobox', { name: 'Dates' }), 'ended')
    expect(screen.getByText('No contests match these filters.')).toBeVisible()
    await user.click(screen.getAllByRole('button', { name: 'Clear filters' })[0])
    expect(screen.getByRole('link', { name: 'Autumn in French' })).toBeVisible()
  })

  it('validates registration, saves selected languages, and changes the live contest action', async () => {
    const user = userEvent.setup()
    open('/contests/novels/registration?scenario=member')
    await user.click(screen.getByRole('button', { name: 'Save registration' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Choose at least one option.')
    const combo = screen.getByRole('combobox', { name: /Your languages/ })
    await user.click(combo)
    await user.type(combo, 'Japanese')
    await user.click(await screen.findByRole('option', { name: 'Japanese' }))
    expect(screen.getByRole('button', { name: 'Remove Japanese' })).toBeInTheDocument()
    combo.focus()
    await user.keyboard('{Escape}')
    expect(screen.getByRole('button', { name: 'Remove Japanese' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Save registration' }))
    expect(await screen.findByText('Registration saved')).toBeVisible()
    expect(screen.getByText(/Your languages: Japanese/)).toBeVisible()
    expect(screen.getByRole('link', { name: 'Log activity' })).toHaveAttribute('href', '/logs/new?contest=novels&asOf=2026-09-05')
  })

  it('lets organizers change collections and create an unlisted contest after fixing invalid dates', async () => {
    const user = userEvent.setup()
    open('/contests/mine?scenario=creator')
    expect(screen.getByRole('link', { name: 'Our Sunday reading group' })).toBeVisible()
    await user.click(screen.getByRole('link', { name: 'Community' }))
    expect(screen.queryByText('Our Sunday reading group')).not.toBeInTheDocument()
    await user.click(screen.getByRole('link', { name: 'Create contest' }))
    await user.type(screen.getByRole('textbox', { name: /Contest title/ }), 'Autumn pages')
    await user.type(screen.getByRole('textbox', { name: /What is this contest about/ }), 'Read a chapter with friends each weekend.')
    const end = screen.getByLabelText(/Ends \(UTC\)/)
    await user.clear(end)
    await user.type(end, '2026-09-01')
    await user.click(screen.getByRole('button', { name: 'Create contest' }))
    expect(await screen.findByText('The end date must be on or after the start date.')).toBeVisible()
    await user.clear(end)
    await user.type(end, '2026-10-31')
    await user.click(screen.getByRole('checkbox', { name: /Unlisted/ }))
    await user.click(screen.getByRole('button', { name: 'Create contest' }))
    expect(await screen.findByRole('heading', { name: 'Autumn pages' })).toBeVisible()
    expect(screen.getByText('Accessible with the link')).toBeVisible()
    await user.click(screen.getByRole('link', { name: 'Manage contest' }))
    expect(screen.getByRole('textbox', { name: /Contest title/ })).toHaveValue('Autumn pages')
    await user.clear(screen.getByRole('textbox', { name: /Contest title/ }))
    await user.type(screen.getByRole('textbox', { name: /Contest title/ }), 'Autumn chapters')
    await user.click(screen.getByRole('button', { name: 'Save changes' }))
    expect(await screen.findByRole('heading', { name: 'Autumn chapters' })).toBeVisible()
  })

  it('does not expose organizer forms to a member or to an organizer editing someone else’s contest', () => {
    const view = open('/contests/new?scenario=member')
    expect(screen.getByRole('heading', { name: 'Organizer access required' })).toBeVisible()
    expect(screen.queryByRole('button', { name: 'Create contest' })).not.toBeInTheDocument()
    view.unmount()
    open('/contests/novels/edit?scenario=creator')
    expect(screen.getByRole('heading', { name: 'Organizer access required' })).toBeVisible()
  })
})
