import { beforeEach, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { initialContests, type SampleContest } from '../src/data'
import { PlaygroundProvider, usePlayground } from '../src/state'
import { ContestEditorPage } from '../src/pages/ContestsPage'

const storageKey = 'tadoku-paper-playground-v1'
beforeEach(() => localStorage.clear())

function ContestState() {
  const { contests } = usePlayground()
  return <output aria-label="Saved contests">{JSON.stringify(contests)}</output>
}
function savedContests(): SampleContest[] {
  return JSON.parse(screen.getByLabelText('Saved contests').textContent!)
}
function persist(contests: SampleContest[]) {
  localStorage.setItem(storageKey, JSON.stringify({ contests, logs: [], joinedContests: [] }))
}
function open(path = '/') {
  return render(<MemoryRouter initialEntries={[path]}><PlaygroundProvider><ContestState /><Routes>
    <Route path="/contests/new" element={<ContestEditorPage />} />
    <Route path="/contests/:contestId/edit" element={<ContestEditorPage />} />
    <Route path="*" element={null} />
  </Routes></PlaygroundProvider></MemoryRouter>)
}

it('adds moderator metadata to older sample contests without replacing saved contest edits', () => {
  const edited = { ...initialContests[0], moderatorIds: undefined, title: 'Our renamed round', start: '2026-09-03', description: 'Keep these edits.', languages: ['French'] }
  const custom = { ...edited, id: 'older-custom-contest', ownerId: 'anton' }
  persist([edited, custom])
  open()
  expect(savedContests()).toEqual([{ ...edited, moderatorIds: initialContests[0].moderatorIds }, custom])
  expect(JSON.parse(localStorage.getItem(storageKey)!).contests).toEqual(savedContests())
})

it('honors saved moderator overrides, including an explicitly empty assignment', () => {
  const contests = [{ ...initialContests[0], moderatorIds: ['noor'] }, { ...initialContests[1], moderatorIds: [] }]
  persist(contests)
  open()
  expect(savedContests()).toEqual(contests)
})

it('explicitly assigns the creator as moderator when a contest is created', async () => {
  const user = userEvent.setup()
  open('/contests/new?scenario=creator')
  await user.type(screen.getByRole('textbox', { name: /Contest title/ }), 'Reading together')
  await user.type(screen.getByRole('textbox', { name: /What is this contest about/ }), 'A shared reading month.')
  await user.click(screen.getByRole('button', { name: 'Create contest' }))
  expect(savedContests().find(contest => contest.title === 'Reading together')).toMatchObject({ ownerId: 'anton', moderatorIds: ['anton'] })
})

it.each([{ moderatorIds: ['mei'] }, { moderatorIds: [] }, { moderatorIds: undefined }])('preserves moderator metadata when editing a contest ($moderatorIds)', async ({ moderatorIds }) => {
  const user = userEvent.setup()
  const contest = { ...initialContests[0], id: 'custom-contest', ownerId: 'anton', moderatorIds }
  persist([contest])
  open('/contests/custom-contest/edit?scenario=creator')
  await user.clear(screen.getByRole('textbox', { name: /Contest title/ }))
  await user.type(screen.getByRole('textbox', { name: /Contest title/ }), 'An updated title')
  await user.click(screen.getByRole('button', { name: 'Save changes' }))
  expect(savedContests()[0]).toMatchObject({ id: contest.id, title: 'An updated title', ownerId: 'anton' })
  expect(savedContests()[0].moderatorIds).toEqual(moderatorIds)
})
