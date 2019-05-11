import 'isomorphic-fetch'

const root = 'http://localhost:4000'

export const get = (endpoint: string) =>
  fetch(`${root}${endpoint}`, {
    method: 'get',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
  })

export const post = (endpoint: string, body: any) =>
  fetch(`${root}${endpoint}`, {
    method: 'post',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })
