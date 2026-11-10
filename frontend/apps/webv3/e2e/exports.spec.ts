import { readFile } from 'node:fs/promises'
import type { APIRequestContext, Page } from '@playwright/test'
import { api, expect, reader, test, webv3 } from './logging-fixtures'
import { apiJson } from './standings'

interface Exported {
  id: string
  language_code: string
  activity_id: number
  amount: number | null
  unit_name: string | null
  duration_seconds: number | null
  score: number
  tags: string[]
  description: string | null
}

const columns = 'id,created_at,language_code,language_name,activity_id,activity_name,amount,unit_name,duration_seconds,score,tags,description'

async function saveLog(request: APIRequestContext, data: Record<string, unknown>) {
  const response = await request.post(`${api}/immersion/logs`, { data: { language_code: 'jpn', tags: [], ...data } })
  expect(response.ok(), `log answered ${response.status()}`).toBeTruthy()
  return (await response.json() as { id: string }).id
}

/** Every active log of the account, read page by page from its public activity. */
async function allLogIds(request: APIRequestContext, userId: string) {
  const ids: string[] = []
  for (let page = 0; ; page++) {
    const { logs, total_size: total } = await apiJson<{ logs: { id: string }[]; total_size: number }>(
      request,
      `/immersion/users/${userId}/logs?page_size=100&page=${page}&include_deleted=false`,
    )
    ids.push(...logs.map(log => log.id))
    if (ids.length >= total || !logs.length) return ids
  }
}

/** Clicks the page's download link and returns the downloaded file's name and text. */
async function download(page: Page, name: string) {
  const link = page.getByRole('link', { name, exact: true })
  await expect(link).toBeVisible()
  const [file] = await Promise.all([page.waitForEvent('download'), link.click()])
  return { filename: file.suggestedFilename(), text: await readFile((await file.path())!, 'utf8') }
}

test('the owner downloads every log as CSV and JSON from an empty year and day, and nobody else gets them', async ({ returningLearner: learner, readerContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const { request } = learner.context
  const stamp = Date.now()
  const formula = await saveLog(request, {
    activity_id: 1, amount: 3, unit_key: 'pages', tags: ['webv3-b', 'webv3-a'], description: `=SUM(${stamp})`,
  })
  const timed = await saveLog(request, { activity_id: 2, duration_seconds: 900, description: `export-listening-${stamp}` })
  const foreign = await saveLog(readerContext.request, { activity_id: 1, amount: 1, unit_key: 'pages', description: `export-foreign-${stamp}` })
  const owned = await allLogIds(request, learner.id)
  expect(owned).toEqual(expect.arrayContaining([formula, timed]))

  const page = await learner.context.newPage()
  // The downloads are offered on a year and on a day that hold no logs, and still export everything.
  await page.goto(`/profile/${learner.id}/statistics/2000`)
  await expect(page.getByRole('heading', { name: 'No activity recorded in 2000' })).toBeVisible()
  const csv = await download(page, 'Download CSV')
  await page.goto(`/profile/${learner.id}/updates/1?date=2000-01-01`)
  const json = await download(page, 'Download JSON')
  await testInfo.attach('export.csv', { body: csv.text, contentType: 'text/csv' })
  await testInfo.attach('export.json', { body: json.text, contentType: 'application/json' })

  expect(json.filename).toBe('tadoku-logs.json')
  const { logs } = JSON.parse(json.text) as { logs: Exported[] }
  expect(logs.map(log => log.id).sort()).toEqual([...owned].sort())
  expect(logs.find(log => log.id === formula)).toMatchObject({
    activity_id: 1, amount: 3, duration_seconds: null, tags: ['webv3-a', 'webv3-b'], description: `=SUM(${stamp})`,
  })
  expect(logs.find(log => log.id === timed)).toMatchObject({ activity_id: 2, duration_seconds: 900, description: `export-listening-${stamp}` })

  expect(csv.filename).toBe('tadoku-logs.csv')
  expect(csv.text.startsWith('﻿')).toBe(true)
  const lines = csv.text.slice(1).trimEnd().split('\n')
  expect(lines[0]).toBe(columns)
  expect(lines).toHaveLength(owned.length + 1)
  const row = lines.find(line => line.startsWith(`${formula},`))
  // CSV joins sorted tags and defuses a description a spreadsheet would read as a formula.
  expect(row).toContain(',"webv3-a, webv3-b",')
  expect(row).toMatch(new RegExp(`,'=SUM\\(${stamp}\\)$`))
  expect(csv.text).not.toContain(foreign)

  // Another account's export holds only its own logs, and a visitor gets none.
  const other = await readerContext.request.get(`${api}/immersion/logs/export?format=json`)
  expect(other.status()).toBe(200)
  const otherIds = (await other.json() as { logs: Exported[] }).logs.map(log => log.id)
  expect(otherIds).toContain(foreign)
  expect(otherIds).not.toContain(formula)
  const anonymous = await page.context().browser()!.newContext()
  expect((await anonymous.request.get(`${api}/immersion/logs/export?format=csv`)).status()).toBe(401)
  await anonymous.close()
})
