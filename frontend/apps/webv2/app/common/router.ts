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
