import { browserAppUrls } from './app-urls'

/** Where browser code calls generated Tadoku API operations. */
export const browserApi = () => ({ baseUrl: browserAppUrls().apiEndpoint })

/**
 * The validated data of a generated Tadoku API call. API errors throw an Error whose message is the HTTP status,
 * which the applications' retry and session handling read; transport and validation failures throw `unavailable`.
 */
export async function apiData<T>(
  request:
    | { data?: T; error?: unknown; response?: Response }
    | Promise<{ data?: T; error?: unknown; response?: Response }>,
): Promise<T> {
  const { data, error, response } = await request
  if (response?.ok && error === undefined) return data as T
  throw new Error(response && !response.ok ? response.status.toString() : 'unavailable')
}

/** `T` with fields made required that the contract leaves optional because request bodies or drafts share its schema. */
export type WithRequired<T, K extends keyof T> = T & { [P in K]-?: NonNullable<T[P]> }

/** The record, failing like a malformed answer when it lacks a field its screen needs. */
export function requireFields<T, K extends keyof T>(record: T, ...fields: K[]): WithRequired<T, K> {
  if (fields.some(field => record[field] === undefined || record[field] === null)) throw new Error('unavailable')
  return record as WithRequired<T, K>
}
