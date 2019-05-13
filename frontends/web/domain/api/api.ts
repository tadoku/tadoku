import 'isomorphic-fetch'
import { getAuthenticationToken } from '../../app/session/storage'

const root = 'http://localhost:4000'

interface APIOptions {
  authenticated?: boolean
}

export const get = (endpoint: string, options?: APIOptions) => {
  let requestOptions: RequestInit = {
    method: 'get',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
  }

  if (options && options.authenticated) {
    requestOptions.headers = {
      ...requestOptions.headers,
      authorization: `Bearer ${getAuthenticationToken()}`,
    }
  }

  return fetch(`${root}${endpoint}`, requestOptions)
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
