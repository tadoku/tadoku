import { User, RawUser } from '../user/interfaces'
import { post } from '../../domain/api/api'

interface SignInResponse {
  token: string
  user: User
}

interface rawSignInResponse {
  token: string
  user: RawUser
}

const signIn = async (payload: {
  email: string
  password: string
}): Promise<SignInResponse | undefined> => {
  const response = await post(`/login`, payload)

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

const register = async (payload: {
  email: string
  password: string
  displayName: string
}): Promise<boolean> => {
  const response = await post(`/register`, payload)

  return response.status === 201
}

const SessionApi = {
  signIn,
  register,
}

export default SessionApi
