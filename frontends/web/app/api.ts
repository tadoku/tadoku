import 'isomorphic-fetch'
import {
  getAuthenticationToken,
  removeUserFromLocalStorage,
} from './session/storage'

// TODO: move this endpoint into env
const root =
  process.env.NODE_ENV === 'production'
    ? 'https://api.tadoku.app'
    : 'http://localhost:4000'

interface APIOptions {
  authenticated?: boolean
}

interface APIOptionsForPost extends APIOptions {
  body: any
}

const request = (
  method: string,
  endpoint: string,
  options: APIOptions | APIOptionsForPost | undefined,
) => {
  const requestOptions: RequestInit = {
    method,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    ...(method === 'post' || method === 'put'
      ? {
          body: JSON.stringify((options as APIOptionsForPost).body),
        }
      : {}),
  }

  if (options && options.authenticated) {
    requestOptions.headers = {
      ...requestOptions.headers,
      authorization: `Bearer ${getAuthenticationToken()}`,
    }
  }

  return fetch(`${root}${endpoint}`, requestOptions)
}

export const get = (endpoint: string, options?: APIOptions) =>
  request('get', endpoint, options).then(response => {
    if (response.status === 401) {
      removeUserFromLocalStorage()
      location.reload()
    }

    return response
  })

export const destroy = (endpoint: string, options?: APIOptions) =>
  request('delete', endpoint, options)

export const post = (endpoint: string, options: APIOptionsForPost) =>
  request('post', endpoint, options)

export const put = (endpoint: string, options: APIOptionsForPost) =>
  request('put', endpoint, options)
