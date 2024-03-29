import {
  SelfServiceRecoveryFlow,
  SubmitSelfServiceRecoveryFlowBody,
} from '@ory/client'
import type { NextPage } from 'next'
import { useEffect, useState } from 'react'
import Flow from '../src/ui/Flow'
import ory from '../src/ory'
import { AxiosError } from 'axios'
import { useSession } from '../src/session'
import { useRouter } from 'next/router'
import { handleFlowError } from '../src/errors'
import { ErrorFallback, withOryErrorBoundary } from '../src/OryErrorBoundary'
import Link from 'next/link'

interface Props {}

const AccountRecovery: NextPage<Props> = () => {
  const [flow, setFlow] = useState<SelfServiceRecoveryFlow>()
  const [session, setSession] = useSession()
  const router = useRouter()
  const { flow: flowId, return_to: returnTo } = router.query
  const [error, setError] = useState<Error>()

  useEffect(() => {
    // Skip if we aren't ready
    if (!router.isReady || flow || error) {
      return
    }

    if (session) {
      router.replace('/')
      return
    }

    // If ?flow=.. was in the URL, we fetch it
    if (flowId) {
      ory
        .getSelfServiceRecoveryFlow(String(flowId))
        .then(({ data }) => {
          setFlow(data)
        })
        .catch(handleFlowError(router, 'recovery', setFlow))
      return
    }

    ory
      .initializeSelfServiceRecoveryFlowForBrowsers()
      .then(({ data }) => {
        console.log(data)
        setFlow(data)
      })
      .catch(handleFlowError(router, 'recovery', setFlow))
      .catch(err => setError(err))
  }, [flowId, router, router.isReady, returnTo, flow, error])

  if (error) {
    return (
      <ErrorFallback
        error={error}
        resetErrorBoundary={() => setError(undefined)}
      />
    )
  }

  if (!flow) {
    return null
  }

  const onSubmit = async (data: SubmitSelfServiceRecoveryFlowBody) => {
    await router.push(`/account-recovery?flow=${flow?.id}`, undefined, {
      shallow: true,
    })

    ory
      .submitSelfServiceRecoveryFlow(flow.id, data)
      .then(async ({ data }) => {
        setFlow(data)
      })
      .catch(handleFlowError(router, 'recovery', setFlow))
      .catch(async (err: AxiosError) => {
        // If the previous handler did not catch the error it's most likely a form validation error
        if (err.response?.status === 400) {
          // Yup, it is!
          setFlow(err.response?.data as SelfServiceRecoveryFlow | undefined)
          return
        }

        return Promise.reject(err)
      })
  }

  return (
    <div>
      <h1 className="title mb-4">Account recovery</h1>
      <div className="card">
        <Flow flow={flow} onSubmit={onSubmit} />
      </div>

      <div className="h-stack items-center space-x-2 mt-4 justify-center text-xs">
        <span className="text-gray-500">Remember your login details?</span>
        <Link href="/login" className="btn ghost small">
          Log in now
        </Link>
      </div>
    </div>
  )
}

export default withOryErrorBoundary(AccountRecovery)
