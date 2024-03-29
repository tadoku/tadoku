import { atom, useAtom } from 'jotai'
import { Session } from '@ory/client'
import { useEffect, DependencyList, useState } from 'react'
import { AxiosError } from 'axios'
import ory from './ory'
import { AppContext } from 'next/app'
import { NextPageContext } from 'next'
import { useRouter } from 'next/router'

export const sessionAtom = atom(undefined as undefined | Session)

export const useSession = () => {
  return useAtom(sessionAtom)
}

// Returns a function which will log the user out
// TODO: cache result as this is now triggering four logout flows at once
export const useLogoutHandler = (deps?: DependencyList) => {
  const [logoutToken, setLogoutToken] = useState<string>('')
  const router = useRouter()

  useEffect(() => {
    ory
      .createSelfServiceLogoutFlowUrlForBrowsers()
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
  }, deps)

  return () => {
    if (logoutToken) {
      ory
        .submitSelfServiceLogoutFlow(logoutToken)
        .then(() => router.push('/login'))
        .then(() => router.reload())
    }
  }
}

// Used to prevent access to a page when a user is not authenticated
export const useProtectedRoute = (fallback: string = '/') => {
  const [session] = useSession()
  const router = useRouter()

  useEffect(() => {
    if (!session) {
      router.replace(fallback)
    }
  }, [session])
}

// Used to prevent access to a page when a user is authenticated
export const useAnonymouseRoute = (fallback: string = '/') => {
  const [session] = useSession()
  const router = useRouter()

  useEffect(() => {
    if (session) {
      router.replace(fallback)
    }
  }, [session])
}

export interface AppContextWithSession extends AppContext {
  ctx: NextPageContextWithSession
}

export interface NextPageContextWithSession extends NextPageContext {
  session: Session | undefined
}
