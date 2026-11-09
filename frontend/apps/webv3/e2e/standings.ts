import type { APIRequestContext, Page } from '@playwright/test'
import { api, expect } from './logging-fixtures'

export interface Entry { rank: number; user_id: string; user_display_name: string; score: number; is_tie: boolean }
export interface Board { entries: Entry[]; total_size: number }

export const score = new Intl.NumberFormat('en-US', { minimumFractionDigits: 1, maximumFractionDigits: 1 })
const count = new Intl.NumberFormat('en-GB', { maximumFractionDigits: 0 })

/** The standings table's participant rows: its header row holds only column headers. */
export const standingRows = (page: Page) => page.getByRole('table').getByRole('row').filter({ has: page.getByRole('cell') })

export async function apiJson<T>(request: APIRequestContext, path: string): Promise<T> {
  const response = await request.get(`${api}${path}`)
  expect(response.ok(), `${path} answered ${response.status()}`).toBeTruthy()
  return response.json() as Promise<T>
}

export const board = (request: APIRequestContext, path: string, page = 1) =>
  apiJson<Board>(request, `${path}?page_size=50&page=${page - 1}`)

/** Every API entry is a row with its rank (T for ties), its name linking to `profile` and its score. */
export async function expectRows(page: Page, entries: Entry[], profile: (userId: string) => string) {
  const rows = standingRows(page)
  await expect(rows).toHaveCount(entries.length)
  for (const [index, entry] of entries.entries()) {
    const row = rows.nth(index)
    await expect(row.getByRole('cell').first()).toHaveText(`${entry.is_tie ? 'T' : ''}${entry.rank}`)
    await expect(row.getByRole('link', { name: entry.user_display_name, exact: true })).toHaveAttribute('href', profile(entry.user_id))
    await expect(row.getByRole('cell').last()).toHaveText(score.format(entry.score))
  }
}

/** The page's count line or empty state agrees with the API's total. */
export async function expectTotal(page: Page, data: Board, pageNumber = 1) {
  if (data.total_size === 0) {
    await expect(page.getByRole('heading', { level: 3 })).toContainText(/No (official scores|participants)/)
    return
  }
  const first = (pageNumber - 1) * 50 + 1
  // The range summary appears only beside page links.
  await expect(page.getByText(/ of [\d,]+ participants$/)).toHaveText(data.total_size > 50 || pageNumber > 1
    ? [`${count.format(first)}–${count.format(first + data.entries.length - 1)} of ${count.format(data.total_size)} participants`]
    : [])
  await expect(page.getByRole('navigation', { name: 'Standings pages' })).toHaveCount(data.total_size > 50 ? 1 : 0)
}
