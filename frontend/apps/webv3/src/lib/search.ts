import { defaultParseSearch, defaultStringifySearch } from '@tanstack/react-router'

/**
 * Search parameters that hold lists of strings, written as repeated keys such as `?tag=a&tag=b`: what a GET form
 * submits and the API reads. Their values stay strings, so a tag such as `2024` isn't read as a number.
 */
const listKeys = ['language', 'activity', 'tag']

export function parseSearch(search: string) {
  const parsed: Record<string, unknown> = defaultParseSearch(search)
  const params = new URLSearchParams(search)
  for (const key of listKeys) if (params.has(key)) parsed[key] = params.getAll(key)
  return parsed
}

export function stringifySearch(search: Record<string, unknown>) {
  const rest = { ...search }
  const lists = new URLSearchParams()
  for (const key of listKeys) {
    const values = rest[key]
    if (!Array.isArray(values)) continue
    delete rest[key]
    for (const value of values) lists.append(key, String(value))
  }
  const query = [defaultStringifySearch(rest).slice(1), lists.toString()].filter(Boolean).join('&')
  return query ? `?${query}` : ''
}
