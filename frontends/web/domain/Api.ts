import { User, RawUser } from './user'
import 'isomorphic-fetch'

const endpoint = 'http://localhost:4000'

const post = (endpoint: string, body: any) =>
  fetch(endpoint, {
    method: 'post',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })

interface SignInResponse {
  token: string
  user: User
}

interface rawSignInResponse {
  token: string
  user: RawUser
}

export const SignIn = async (payload: {
  email: string
  password: string
}): Promise<SignInResponse | undefined> => {
  const response = await post(`${endpoint}/login`, payload)

  if (response.status !== 200) {
    return undefined
  }

  const data: rawSignInResponse = await response.json()

  return {
    token: data.token,
    user: {
      id: data.user.id,
      email: data.user.email,
      displayName: data.user.display_name,
    },
  }
}

export const Register = async (payload: {
  email: string
  password: string
  displayName: string
}): Promise<boolean> => {
  const response = await post(`${endpoint}/register`, payload)

  return response.status === 201
}
