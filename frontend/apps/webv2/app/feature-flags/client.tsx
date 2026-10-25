import { immersionFeatureFlagDecisions } from 'api-client'
import { atom, useAtomValue, useSetAtom } from 'jotai'
import { browserAppUrls } from 'ui/app-urls'
import { useRouter } from 'next/router'
import React, { useEffect } from 'react'
import type { ReactNode } from 'react'
import { defaultFeatureFlagDecisions, registeredDecisions } from './registry'
import type { FeatureFlagDecisions, FeatureFlagKey } from './registry'

export const featureFlagDecisionsAtom = atom<FeatureFlagDecisions>({
  ...defaultFeatureFlagDecisions,
})

export const useFeatureFlag = (flag: FeatureFlagKey) =>
  useAtomValue(featureFlagDecisionsAtom)[flag]

export const FeatureFlagRefresh = ({
  children,
  timeoutMilliseconds = 3_000,
}: {
  children: ReactNode
  timeoutMilliseconds?: number
}) => {
  const router = useRouter()
  const setDecisions = useSetAtom(featureFlagDecisionsAtom)

  useEffect(() => {
    let controller: AbortController | undefined

    const refresh = () => {
      controller?.abort()
      controller = new AbortController()
      const currentController = controller
      setDecisions({ ...defaultFeatureFlagDecisions })
      const timeout = setTimeout(
        () => currentController.abort(),
        timeoutMilliseconds,
      )

      void immersionFeatureFlagDecisions({
        baseUrl: browserAppUrls().apiEndpoint,
        credentials: 'include',
        cache: 'no-store',
        signal: currentController.signal,
      })
        .then(({ data }) => {
          if (data) {
            setDecisions(registeredDecisions(data.decisions))
          }
        })
        .finally(() => clearTimeout(timeout))
    }

    router.events.on('routeChangeComplete', refresh)
    return () => {
      controller?.abort()
      router.events.off('routeChangeComplete', refresh)
    }
  }, [router.events, setDecisions, timeoutMilliseconds])

  return <>{children}</>
}
