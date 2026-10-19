import { atom, useAtom } from 'jotai'
import { Account, AccountApp, logout } from 'ui/identity'
import { useRouter } from 'next/router'
import { useEffect, useRef } from 'react'
import { z } from 'zod'
import { browserAppUrls } from 'ui/app-urls'
import { AppContext } from 'next/app'
import { NextPageContext } from 'next'
import { useCurrentLocation } from '@app/common/hooks'
import { useRoutes } from './routes'

const root = () => `${browserAppUrls().apiEndpoint}/authz`

export type Role = 'admin' | 'user' | 'guest' | 'banned'

const UserRoleResponse = z.object({
  role: z.enum(['admin', 'user', 'guest', 'banned']),
})

export const sessionAtom = atom(undefined as undefined | Account)
export const accountAppAtom = atom<AccountApp>('v1')
export const useAccountApp = () => useAtom(accountAppAtom)[0]

export const userRoleAtom = atom(undefined as undefined | Role)

export const useSession = () => {
  return useAtom(sessionAtom)
}

export const useSessionOrRedirect = () => {
  const routes = useRoutes()
  const result = useAtom(sessionAtom)
  const currentUrl = useCurrentLocation()
  const router = useRouter()
  const loginUrl = currentUrl ? routes.authLogin(currentUrl) : undefined

  const [session] = result

  useEffect(() => {
    if (!session && router.isReady && loginUrl) {
      router.push(loginUrl)
    }
  }, [session, router, loginUrl])

  return result
}

export const useLogoutHandler = () => {
  const routes = useRoutes()
  const [, setSession] = useSession()
  return async () => {
    await logout()
    setSession(undefined)
    window.location.assign(routes.authLogin())
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

    fetch(`${root()}/current-user/role`)
      .then(async response => {
        if (response.status !== 200) {
          throw new Error(response.status.toString())
        }
        const data = UserRoleResponse.parse(await response.json())
        setRole(data.role)
      })
      .catch(() => {
        setRole(undefined)
      })
  }, [userId, setRole])

  return role
}

export interface AppContextWithSession extends AppContext {
  ctx: NextPageContextWithSession
}

export interface NextPageContextWithSession extends NextPageContext {
  session: Account | undefined
}
