import 'isomorphic-fetch'
import { getService } from '@app/services'

interface ApiClient {
  get(endpoint: string): ReturnType<typeof fetch>
  destroy(endpoint: string): ReturnType<typeof fetch>
  post(endpoint: string, options: APIOptionsForPost): ReturnType<typeof fetch>
  put(endpoint: string, options: APIOptionsForPost): ReturnType<typeof fetch>
}

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

  return fetch(endpoint, requestOptions)
}

const createGet = (rootUrl: string) => (endpoint: string) =>
  request('get', `${rootUrl}${endpoint}`).then(response => {
    if (response.status === 401) {
      location.reload()
    }

    return response
  })

const createDestroy = (rootUrl: string) => (endpoint: string) =>
  request('delete', `${rootUrl}${endpoint}`)

const createPost =
  (rootUrl: string) => (endpoint: string, options: APIOptionsForPost) =>
    request('post', `${rootUrl}${endpoint}`, options)

const createPut =
  (rootUrl: string) => (endpoint: string, options: APIOptionsForPost) =>
    request('put', `${rootUrl}${endpoint}`, options)

export const createApiClient = (rootUrl: string): ApiClient => ({
  get: createGet(rootUrl),
  destroy: createDestroy(rootUrl),
  post: createPost(rootUrl),
  put: createPut(rootUrl),
})

const defaultApiClient = createApiClient(
  getService('tadokuContest').externalUrl,
)

export const get = defaultApiClient.get
export const post = defaultApiClient.post
export const destroy = defaultApiClient.destroy
export const put = defaultApiClient.put
