import 'isomorphic-fetch'

const root = 'http://localhost:4000'

export const post = (endpoint: string, body: any) =>
  fetch(`${root}${endpoint}`, {
    method: 'post',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })
