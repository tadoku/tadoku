import { User } from './interfaces'

const LOCAL_STORAGE_USER_KEY = 'user?i=2'

export const storeUserInLocalStorage = (user: {
  expiresAt: number
  user: User
}) => {
  window.localStorage.setItem(LOCAL_STORAGE_USER_KEY, JSON.stringify(user))
}

export const removeUserFromLocalStorage = () => {
  window.localStorage.clear()
}

export const loadUserFromLocalStorage = (): {
  expiresAt: number
  user: User
} | null => {
  const user = window.localStorage.getItem(LOCAL_STORAGE_USER_KEY)

  if (!user) {
    return null
  }

  return JSON.parse(user)
}
