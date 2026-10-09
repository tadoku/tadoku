import { DateTime } from 'luxon'

export function getQueryStringIntParameter(
  param: string | string[] | undefined,
  fallback: number,
) {
  if (!param) {
    return fallback
  }

  const parsed = parseInt(param.toString())
  if (isNaN(parsed)) {
    return fallback
  }

  return parsed
}

export function getQueryStringPageParameter(
  param: string | string[] | undefined,
) {
  return Math.max(1, getQueryStringIntParameter(param, 1))
}

export function getQueryStringDateParameter(
  param: string | string[] | undefined,
) {
  const value = param?.toString() ?? ''
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value) || !DateTime.fromISO(value).isValid) {
    return undefined
  }

  return value
}
