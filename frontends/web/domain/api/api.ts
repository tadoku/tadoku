import 'isomorphic-fetch'
import { getAuthenticationToken } from '../Session'

const root = 'http://localhost:4000'

interface APIOptions {
  authenticated?: boolean
}

export const get = (endpoint: string, options?: APIOptions) => {
  const authenticationHeader = {
    authorization: `Bearer ${getAuthenticationToken()}`,
  }

  return fetch(`${root}${endpoint}`, {
    method: 'get',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...((options || {}).authenticated ? authenticationHeader : {}),
    },
  })
}

export const post = (endpoint: string, body: any) =>
  fetch(`${root}${endpoint}`, {
    method: 'post',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })
