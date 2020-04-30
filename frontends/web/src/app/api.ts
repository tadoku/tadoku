import 'isomorphic-fetch'
import { removeUserFromLocalStorage } from './session/storage'

const root = '/api'

interface APIOptionsForPost {
  body: any
}

const request = (
  method: string,
  endpoint: string,
  options?: APIOptionsForPost,
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

  return fetch(`${root}${endpoint}`, requestOptions)
}

export const get = (endpoint: string) =>
  request('get', endpoint).then(response => {
    if (response.status === 401) {
      removeUserFromLocalStorage()
      location.reload()
    }

    return response
  })

export const destroy = (endpoint: string) => request('delete', endpoint)

export const post = (endpoint: string, options: APIOptionsForPost) =>
  request('post', endpoint, options)

export const put = (endpoint: string, options: APIOptionsForPost) =>
  request('put', endpoint, options)
