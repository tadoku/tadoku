import { User } from './User'

export const storeUser = (user: { token: string; user: User }) => {
  window.localStorage.setItem('user', JSON.stringify(user))
}
export const loadUser = (): { token: string; user: User } | null => {
  const user = window.localStorage.getItem('user')

  if (!user) {
    return null
  }

  return JSON.parse(user)
}
