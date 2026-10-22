import { authzRoleGet, type AuthzUserRole } from 'api-client'
import { atom, useAtom } from 'jotai'
import { toast } from 'react-toastify'
import { Account, logout } from 'ui/identity'
import { useRouter } from 'next/router'
import { useEffect, useRef } from 'react'
import { browserApi } from 'ui/api'
import { AppContext } from 'next/app'
import { NextPageContext } from 'next'
import { useRoutes } from './routes'

export type Role = AuthzUserRole['role']

export const sessionAtom = atom(undefined as undefined | Account)

export const userRoleAtom = atom(undefined as undefined | Role)

export const useSession = () => {
  return useAtom(sessionAtom)
}

export const useCurrentLocation = () => {
  if (typeof window === 'undefined') {
    return ''
  }
  return window.location.href
}

export const useSessionOrRedirect = () => {
  const routes = useRoutes()
  const result = useAtom(sessionAtom)
  const currentUrl = useCurrentLocation()
  const router = useRouter()

  const [session] = result

  useEffect(() => {
    if (!session) {
      router.push(routes.authLogin(currentUrl))
    }
  })

  return result
}

export const useLogoutHandler = () => {
  const routes = useRoutes()
  const [, setSession] = useSession()
  return async () => {
    try {
      await logout()
      setSession(undefined)
      window.location.assign(routes.authLogin())
    } catch {
      toast.error('Log out failed. Please try again.')
    }
  }
}

export const useUserRole = () => {
  const [session] = useAtom(sessionAtom)
  const [role, setRole] = useAtom(userRoleAtom)
  const prevUserIdRef = useRef<string | undefined>(undefined)

  const userId = session?.id

  useEffect(() => {
    if (!userId) {
      prevUserIdRef.current = undefined
      setRole(undefined)
      return
    }

    if (userId === prevUserIdRef.current) {
      return
    }

    prevUserIdRef.current = userId

    void authzRoleGet(browserApi()).then(({ data }) => setRole(data?.role))
  }, [userId, setRole])

  return role
}

export interface AppContextWithSession extends AppContext {
  ctx: NextPageContextWithSession
}

export interface NextPageContextWithSession extends NextPageContext {
  session: Account | undefined
}
