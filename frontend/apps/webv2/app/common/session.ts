import { atom, useAtom } from 'jotai'
import { Session } from '@ory/client'
import { useRouter } from 'next/router'
import { useEffect, DependencyList } from 'react'
import { AxiosError } from 'axios'
import ory from '@app/common/ory'
import { AppContext } from 'next/app'
import { NextPageContext } from 'next'
import getConfig from 'next/config'
import { useCurrentLocation } from '@app/common/hooks'

const { publicRuntimeConfig } = getConfig()

export const sessionAtom = atom(undefined as undefined | Session)

export const useSession = () => {
  return useAtom(sessionAtom)
}

export const useSessionOrRedirect = () => {
  const result = useAtom(sessionAtom)
  const currentUrl = useCurrentLocation()
  const router = useRouter()

  const [session] = result

  useEffect(() => {
    if (!session) {
      router.push(
        publicRuntimeConfig.authUiUrl + `/login?return_to=${currentUrl}`,
      )
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
        .then(() => router.push('/login'))
        .then(() => router.reload())
    }
  }
}

export interface AppContextWithSession extends AppContext {
  ctx: NextPageContextWithSession
}

export interface NextPageContextWithSession extends NextPageContext {
  session: Session | undefined
}
