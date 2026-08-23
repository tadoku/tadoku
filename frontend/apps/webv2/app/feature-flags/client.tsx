import { atom, useAtomValue, useSetAtom } from 'jotai'
import { useRouter } from 'next/router'
import React, { useEffect } from 'react'
import type { ReactNode } from 'react'
import {
  defaultFeatureFlagDecisions,
  featureFlagResponseSchema,
} from './registry'
import type { FeatureFlagDecisions, FeatureFlagKey } from './registry'

export const featureFlagDecisionsAtom = atom<FeatureFlagDecisions>({
  ...defaultFeatureFlagDecisions,
})

export const useFeatureFlag = (flag: FeatureFlagKey) =>
  useAtomValue(featureFlagDecisionsAtom)[flag]

export const FeatureFlagRefresh = ({ children }: { children: ReactNode }) => {
  const router = useRouter()
  const setDecisions = useSetAtom(featureFlagDecisionsAtom)

  useEffect(() => {
    let controller: AbortController | undefined

    const refresh = () => {
      controller?.abort()
      controller = new AbortController()

      void fetch('/api/feature-flags', {
        credentials: 'same-origin',
        cache: 'no-store',
        signal: controller.signal,
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
    }

    router.events.on('routeChangeComplete', refresh)
    return () => {
      controller?.abort()
      router.events.off('routeChangeComplete', refresh)
    }
  }, [router.events, setDecisions])

  return <>{children}</>
}
