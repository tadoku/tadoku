import { User } from './interfaces'

const LOCAL_STORAGE_USER_KEY = 'user'

export const storeUserInLocalStorage = (user: {
  token: string
  user: User
}) => {
  window.localStorage.setItem(LOCAL_STORAGE_USER_KEY, JSON.stringify(user))
}

export const removeUserFromLocalStorage = () => {
  window.localStorage.removeItem(LOCAL_STORAGE_USER_KEY)
}

export const loadUserFromLocalStorage = (): {
  token: string
  user: User
} | null => {
  const user = window.localStorage.getItem(LOCAL_STORAGE_USER_KEY)

  if (!user) {
    return null
  }

  return JSON.parse(user)
}

export const getAuthenticationToken = (): string | null => {
  const user = loadUserFromLocalStorage()

  if (!user) {
    return null
  }

  return user.token
}
