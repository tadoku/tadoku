import type { FilterMenuSection, FilterMenuValue } from 'paper-ui'
import { z } from 'zod'
import { activityId, activitySlug, type ActivitySlug } from './activity'
import type { ContestSummary } from './contest-activity'

/**
 * A contest profile's filters as its address holds them: language codes, lower-case activity names such as
 * `listening`, and tags. Values in one filter match any of them; filters combine.
 */
export type ProfileFilters = { language: string[]; activity: ActivitySlug[]; tag: string[] }

export const noFilters: ProfileFilters = { language: [], activity: [], tag: [] }

/** The filters as the API reads them, with the contract's limits: language codes, activity IDs and tags. */
export const logFiltersSchema = z.object({
  language: z.array(z.string().min(1).max(10)).max(10),
  activity: z.array(z.int().min(1).max(32767)).max(10),
  tag: z.array(z.string().refine(tag => tag.length > 0 && [...tag].length <= 50)).max(20),
})

export type LogFilters = z.infer<typeof logFiltersSchema>

const values = (raw: unknown) =>
  (Array.isArray(raw) ? raw : [raw]).flatMap(value => (typeof value === 'string' ? [value.trim()] : []))

const distinct = (items: string[], limit: number) => [...new Set(items)].slice(0, limit)

/**
 * Reads the filters from an address's search, dropping what the API would refuse or can't match: malformed language
 * codes, unknown activities, empty or overlong tags, repeats and values past each filter's limit.
 */
export function parseFilters(search: Record<string, unknown>): ProfileFilters {
  return {
    language: distinct(values(search.language).filter(code => /^[\w-]{1,10}$/.test(code)), 10),
    activity: distinct(values(search.activity).filter(name => activityId(name) !== undefined), 10) as ActivitySlug[],
    tag: distinct(values(search.tag).filter(tag => tag.length > 0 && [...tag].length <= 50), 20),
  }
}

export const toLogFilters = (filters: ProfileFilters): LogFilters => ({
  language: filters.language,
  activity: filters.activity.map(name => activityId(name)!),
  tag: filters.tag,
})

export const filterCount = (filters: ProfileFilters) => filters.language.length + filters.activity.length + filters.tag.length

/**
 * The Filters entry point's sections: Language when the participant registered more than one, then Activity and Tag.
 * Each value carries its points under the other filters; a chosen value the breakdown no longer lists stays, at 0.
 */
export function filterSections(summary: ContestSummary | undefined, filters: ProfileFilters, languages: { code: string; name: string }[]): FilterMenuSection[] {
  const languageName = (code: string) => languages.find(language => language.code === code)?.name ?? code
  const section = (id: keyof ProfileFilters, label: string, values: FilterMenuValue[], name = (value: string) => value): FilterMenuSection => ({
    id,
    label,
    selected: filters[id],
    values: [
      ...values,
      ...filters[id].filter(value => !values.some(item => item.value === value)).map(value => ({ value, label: name(value), points: 0 })),
    ],
  })
  const activities = (summary?.activities ?? []).flatMap(item => {
    const slug = activitySlug(item.id)
    return slug ? [{ value: slug, label: item.name, points: item.score }] : []
  })

  return [
    ...(languages.length > 1
      ? [section('language', 'Language', (summary?.languages ?? []).map(item => ({ value: item.code, label: languageName(item.code), points: item.score })), languageName)]
      : []),
    section('activity', 'Activity', activities, capitalized),
    section('tag', 'Tag', (summary?.tags ?? []).map(item => ({ value: item.tag, label: item.tag, points: item.score }))),
  ]
}

const capitalized = (value: string) => value.charAt(0).toUpperCase() + value.slice(1)
