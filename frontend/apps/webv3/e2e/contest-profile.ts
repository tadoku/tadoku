import { randomUUID } from 'node:crypto'
import type { APIRequestContext, BrowserContext, Locator, Page } from '@playwright/test'
import { api, expect, statValue } from './logging-fixtures'
import { apiJson, score } from './standings'

export interface Contest { id: string; title: string; contest_start: string; contest_end: string }
export interface Log {
  id: string
  description?: string
  tags: string[]
  score: number
  language: { name: string }
  activity: { name: string }
  user_display_name?: string
}
export interface Logs { logs: Log[]; total_size: number }
export interface Summary {
  score: number
  log_count: number
  active_days: number
  best_day: { date: string; score: number } | null
  days: { date: string; activity_id: number; score: number }[]
  languages: { code: string; name: string; score: number }[]
  activities: { id: number; name: string; score: number }[]
  tags: { tag: string; score: number; log_count: number }[]
}

/** Filters as the API reads them: language codes, activity IDs and tags. */
export interface Filters { language?: string[]; activity?: number[]; tag?: string[] }

export const day = (offset = 0) => new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10)
export const profilePath = (id: string, userId: string, page = 1) => `/contests/${id}/profile/${userId}/${page}`
export const fitsViewport = (page: Page) => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)

/** Repeated search parameters, as the API and the page's address both read filters. */
export function filterQuery(filters: Filters = {}) {
  const query = new URLSearchParams()
  for (const [key, values] of Object.entries(filters)) for (const value of values ?? []) query.append(key, String(value))
  return query
}

/** One page of a contest's logs, counted from 1, as the API lists them. */
export function contestLogs(request: APIRequestContext, contestId: string, page = 1, userId?: string, filters?: Filters) {
  const query = new URLSearchParams({ page_size: '25', page: String(page - 1), include_deleted: 'false', ...(userId ? { user_id: userId } : {}) })
  for (const [key, value] of filterQuery(filters)) query.append(key, value)
  return apiJson<Logs>(request, `/immersion/contests/${contestId}/logs?${query}`)
}

export const profileSummary = (request: APIRequestContext, contestId: string, userId: string, filters?: Filters) =>
  apiJson<Summary>(request, `/immersion/contests/${contestId}/profile/${userId}/summary?${filterQuery(filters)}`)

/** A contest created through the API as the context's account, open today, accepting reading and listening. */
export async function createContest(context: BrowserContext, days = 30, title = `Webv3 moderation check ${randomUUID().slice(0, 8)}`): Promise<Contest> {
  const response = await context.request.post(`${api}/immersion/contests`, {
    data: {
      title,
      description: 'Created by the Webv3 contest journeys.',
      contest_start: day(),
      contest_end: day(days),
      registration_end: day(days - 16),
      private: false,
      official: false,
      language_code_allow_list: null,
      activity_type_id_allow_list: [1, 2],
    },
  })
  expect(response.ok(), `contest creation answered ${response.status()}`).toBeTruthy()
  return { ...(await response.json()), title }
}

/** Registers the context's account for the contest in these languages and returns the registration's ID. */
export async function register(context: BrowserContext, contestId: string, languages = ['jpn']) {
  const joined = await context.request.post(`${api}/immersion/contests/${contestId}/registration`, { data: { language_codes: languages } })
  expect(joined.status()).toBe(200)
  return (await apiJson<{ id: string }>(context.request, `/immersion/contests/${contestId}/registration`)).id
}

export type NewLog =
  | { language: string; activity: 2; minutes: number; tags: string[] }
  | { language: string; activity: 1; pages: number; tags: string[] }

/** Submits one log to a registration through the API. */
export async function createLog(context: BrowserContext, registrationId: string, log: NewLog, description: string) {
  const amount = log.activity === 1 ? { unit_key: 'pages', amount: log.pages } : { duration_seconds: log.minutes * 60 }
  const created = await context.request.post(`${api}/immersion/logs`, {
    data: { language_code: log.language, activity_id: log.activity, ...amount, tags: log.tags, description, registration_ids: [registrationId] },
  })
  expect(created.status(), await created.text()).toBe(200)
  return (await created.json()) as Log
}

/** Registers the context's account for the contest and submits one listening log to it through the API. */
export async function submitLog(context: BrowserContext, contestId: string, description: string) {
  const registration = await register(context, contestId)
  return createLog(context, registration, { language: 'jpn', activity: 2, minutes: 10, tags: [] }, description)
}

const escape = (text: string) => text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

/** The summary's four figures, and its chart's description, agree with the API's contest-profile summary. */
export async function expectSummary(page: Page, summary: Summary) {
  await expect(statValue(page, 'Score')).toHaveText(new RegExp(`^${escape(score.format(summary.score))}`))
  await expect(statValue(page, 'Active days')).toHaveText(new RegExp(`^${summary.active_days}`))
  await expect(statValue(page, 'Logs')).toHaveText(String(summary.log_count))
  await expect(statValue(page, 'Best day')).toHaveText(new RegExp(`^${summary.best_day ? escape(score.format(summary.best_day.score)) : '–'}`))
  await expect(activityChart(page)).toHaveAccessibleDescription(new RegExp(`^${escape(score.format(summary.score))} points`))
}

export const activityChart = (page: Page) => page.getByRole('img', { name: 'Points per day by activity, with the running total' })

/** A breakdown list's rows, by label and points, in the order the page shows them. */
export const breakdownRows = (list: Locator) => list.getByRole('listitem')

/** The activity and tag lists show the API's breakdowns, each counted under every other filter. */
export async function expectBreakdowns(page: Page, summary: Summary) {
  const activities = breakdownRows(page.getByRole('region', { name: 'Score by activity' }))
  await expect(activities).toHaveText(summary.activities.map(activity => new RegExp(`^${escape(activity.name)}\\s*${escape(score.format(activity.score))} point`)))
  const tags = breakdownRows(page.getByRole('region', { name: 'Score by tag' }))
  await expect(tags).toHaveText(summary.tags.slice(0, 6).map(tag =>
    new RegExp(`^${escape(tag.tag)}\\s*${escape(score.format(tag.score))} points? · ${tag.log_count} logs?$`)))
}
