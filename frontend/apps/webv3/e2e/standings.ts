import type { APIRequestContext, Page } from '@playwright/test'
import { api, expect } from './logging-fixtures'

export interface Entry { rank: number; user_id: string; user_display_name: string; score: number; is_tie: boolean }
export interface Board { entries: Entry[]; total_size: number }

export const score = new Intl.NumberFormat('en-GB', { maximumFractionDigits: 1 })

/** The page's single-decimal score column: one decimal on every row once any row has one. */
const scoreColumn = (entries: Entry[]) => {
  const digits = entries.some(entry => Math.round(entry.score * 10) % 10 !== 0) ? 1 : 0
  return new Intl.NumberFormat('en-GB', { minimumFractionDigits: digits, maximumFractionDigits: digits })
}

export async function apiJson<T>(request: APIRequestContext, path: string): Promise<T> {
  const response = await request.get(`${api}${path}`)
  expect(response.ok(), `${path} answered ${response.status()}`).toBeTruthy()
  return response.json() as Promise<T>
}

export const board = (request: APIRequestContext, path: string, page = 1) =>
  apiJson<Board>(request, `${path}?page_size=50&page=${page - 1}`)

/** Every API entry is a row with its rank (T for ties), its name linking to `profile` and its score. */
export async function expectRows(page: Page, entries: Entry[], profile: (userId: string) => string) {
  const rows = page.locator('.standings-table tbody tr')
  const column = scoreColumn(entries)
  await expect(rows).toHaveCount(entries.length)
  for (const [index, entry] of entries.entries()) {
    const row = rows.nth(index)
    await expect(row.locator('td').first()).toHaveText(`${entry.is_tie ? 'T' : ''}${entry.rank}`)
    await expect(row.getByRole('link', { name: entry.user_display_name, exact: true })).toHaveAttribute('href', profile(entry.user_id))
    await expect(row.locator('td').last()).toHaveText(column.format(entry.score))
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
  await expect(page.locator('.standings-pagination > p')).toHaveText(data.total_size > 50 || pageNumber > 1
    ? [`${score.format(first)}–${score.format(first + data.entries.length - 1)} of ${score.format(data.total_size)} participants`]
    : [])
  await expect(page.getByRole('navigation', { name: 'Standings pages' })).toHaveCount(data.total_size > 50 ? 1 : 0)
}
