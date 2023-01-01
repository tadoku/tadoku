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
