import { atom, useAtom } from 'jotai'
import { Session } from '@ory/client'
import { useRouter } from 'next/router'
import { useEffect, useRef, DependencyList } from 'react'
import { AxiosError } from 'axios'
import { z } from 'zod'
import getConfig from 'next/config'
import ory from '@app/common/ory'
import { AppContext } from 'next/app'
import { NextPageContext } from 'next'
import { routes } from './routes'

const { publicRuntimeConfig } = getConfig()
const root = `${publicRuntimeConfig.apiEndpoint}/immersion`

export type Role = 'admin' | 'user' | 'guest' | 'banned'

const UserRoleResponse = z.object({
  role: z.enum(['admin', 'user', 'guest', 'banned']),
})

export const sessionAtom = atom(undefined as undefined | Session)
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

export const logoutTokenAtom = atom(undefined as undefined | string)

// Returns a function which will log the user out
export const useLogoutHandler = (deps?: DependencyList) => {
  const [logoutToken, setLogoutToken] = useAtom(logoutTokenAtom)
  const [session] = useSession()
  const router = useRouter()

  useEffect(() => {
    if (logoutToken || !session) {
      return
    }

    ory
      .createSelfServiceLogoutFlowUrlForBrowsers(undefined, {
        withCredentials: true,
      })
      .then(({ data }) => {
        setLogoutToken(data.logout_token)
      })
      .catch((err: AxiosError) => {
        switch (err.response?.status) {
          case 401:
            // do nothing, the user is not logged in
            return
        }

        // Something else happened!
        return Promise.reject(err)
      })
  }, [...(deps ?? []), logoutToken, session])

  return () => {
    if (logoutToken) {
      ory
        .submitSelfServiceLogoutFlow(logoutToken, undefined, {
          withCredentials: true,
        })
        .then(() => router.push(routes.authLogin()))
        .then(() => router.reload())
    }
  }
}

export const useUserRole = () => {
  const [session] = useAtom(sessionAtom)
  const [role, setRole] = useAtom(userRoleAtom)
  const prevUserIdRef = useRef<string | undefined>(undefined)

  const userId = session?.identity?.id

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

    fetch(`${root}/current-user/role`, { credentials: 'include' })
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
  session: Session | undefined
}
