import { User, RawUser } from './interfaces'
import { post, destroy } from '../api'

interface LogInResponse {
  expiresAt: number
  user: User
}

interface RawLogInResponse {
  expiresAt: number
  user: RawUser
}

const logIn = async (payload: {
  email: string
  password: string
}): Promise<LogInResponse | undefined> => {
  const response = await post(`/sessions`, { body: payload })

  if (response.status !== 200) {
    return undefined
  }

  const data: RawLogInResponse = await response.json()

  return {
    expiresAt: data.expiresAt,
    user: {
      id: data.user.id,
      email: data.user.email,
      displayName: data.user.display_name,
      role: data.user.role,
    },
  }
}

const logOut = async (): Promise<boolean> => {
  const response = await destroy(`/sessions`)
  return response.status === 200
}

const refresh = async (): Promise<LogInResponse | undefined> => {
  const response = await post(`/sessions/refresh`, {
    body: {},
  })

  if (response.status !== 200) {
    return undefined
  }

  const data: RawLogInResponse = await response.json()

  return {
    expiresAt: data.expiresAt,
    user: {
      id: data.user.id,
      email: data.user.email,
      displayName: data.user.display_name,
      role: data.user.role,
    },
  }
}

const register = async (payload: {
  email: string
  password: string
  displayName: string
}): Promise<boolean> => {
  const response = await post(`/users`, {
    body: {
      email: payload.email,
      password: payload.password,
      display_name: payload.displayName,
    },
  })

  return response.status === 201
}

const SessionApi = {
  logIn,
  logOut,
  refresh,
  register,
}

export default SessionApi
