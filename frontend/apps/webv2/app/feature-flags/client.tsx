import { atom, useAtomValue, useSetAtom } from 'jotai'
import getConfig from 'next/config'
import { useRouter } from 'next/router'
import React, { useEffect } from 'react'
import type { ReactNode } from 'react'
import {
  defaultFeatureFlagDecisions,
  featureFlagResponseSchema,
} from './registry'
import type { FeatureFlagDecisions, FeatureFlagKey } from './registry'

const { publicRuntimeConfig } = getConfig()
const featureFlagEndpoint = `${publicRuntimeConfig.apiEndpoint}/immersion/feature-flags`

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

      void fetch(featureFlagEndpoint, {
        credentials: 'include',
        cache: 'no-store',
        signal: currentController.signal,
      })
        .then(async response => {
          if (!response.ok) {
            throw new Error(response.status.toString())
          }
          return featureFlagResponseSchema.parse(await response.json())
            .decisions
        })
        .then(setDecisions)
        .catch(() => {})
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
