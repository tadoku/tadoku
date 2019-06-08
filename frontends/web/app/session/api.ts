import { User, RawUser } from './interfaces'
import { post } from '../api'

interface LogInResponse {
  token: string
  user: User
}

interface rawLogInResponse {
  token: string
  user: RawUser
}

const logIn = async (payload: {
  email: string
  password: string
}): Promise<LogInResponse | undefined> => {
  const response = await post(`/login`, { body: payload })

  if (response.status !== 200) {
    return undefined
  }

  const data: rawLogInResponse = await response.json()

  return {
    token: data.token,
    user: {
      id: data.user.id,
      email: data.user.email,
      displayName: data.user.display_name,
    },
  }
}

const register = async (payload: {
  email: string
  password: string
  displayName: string
}): Promise<boolean> => {
  const response = await post(`/register`, {
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
  register,
}

export default SessionApi
