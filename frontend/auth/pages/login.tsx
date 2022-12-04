import {
  SelfServiceLoginFlow,
  SubmitSelfServiceLoginFlowBody,
} from '@ory/client'
import type { NextPage } from 'next'
import { useEffect, useState } from 'react'
import Flow from '../ui/Flow'
import ory from '../src/ory'
import { AxiosError } from 'axios'
import { useLogoutHandler, useSession } from '../src/session'
import { useRouter } from 'next/router'
import { handleFlowError } from '../src/errors'
import Link from 'next/link'
import { ErrorFallback, withOryErrorBoundary } from '../src/OryErrorBoundary'

interface Props {}

const Login: NextPage<Props> = () => {
  const [flow, setFlow] = useState<SelfServiceLoginFlow>()
  const [session, setSession] = useSession()
  const router = useRouter()
  const { flow: flowId, return_to: returnTo, refresh, aal } = router.query
  const [error, setError] = useState<Error>()

  // This might be confusing, but we want to show the user an option
  // to sign out if they are performing two-factor authentication!
  const onLogout = useLogoutHandler([aal, refresh])

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
        .getSelfServiceLoginFlow(String(flowId))
        .then(({ data }) => {
          setFlow(data)
        })
        .catch(handleFlowError(router, 'login', setFlow))
      return
    }

    ory
      .initializeSelfServiceLoginFlowForBrowsers(
        Boolean(refresh),
        aal ? String(aal) : undefined,
        returnTo ? String(returnTo) : undefined,
      )
      .then(({ data }) => {
        console.log(data)
        setFlow(data)
      })
      .catch(handleFlowError(router, 'login', setFlow))
      .catch(err => setError(err))
  }, [flowId, router, router.isReady, aal, refresh, returnTo, flow, error])

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

  const onSubmit = async (data: SubmitSelfServiceLoginFlowBody) => {
    await router.push(`/login?flow=${flow?.id}`, undefined, {
      shallow: true,
    })

    ory
      .submitSelfServiceLoginFlow(flow.id, data)
      .then(async ({ data }) => {
        if (flow?.return_to) {
          window.location.href = flow?.return_to
          return
        }

        const res = await ory.toSession()
        setSession(res.data)

        router.push('/')
      })
      .catch(handleFlowError(router, 'login', setFlow))
      .catch(async (err: AxiosError) => {
        // If the previous handler did not catch the error it's most likely a form validation error
        if (err.response?.status === 400) {
          // Yup, it is!
          setFlow(err.response?.data)
          return
        }

        return Promise.reject(err)
      })
  }

  return (
    <div>
      <h1>
        {(() => {
          if (flow?.refresh) {
            return 'Confirm Action'
          } else if (flow?.requested_aal === 'aal2') {
            return 'Two-Factor Authentication'
          }
          return 'Sign In'
        })()}
      </h1>
      <Flow flow={flow} onSubmit={onSubmit} />

      {aal || refresh ? (
        <a onClick={onLogout}>Log out</a>
      ) : (
        <>
          <Link href="/register">
            <a>Create account</a>
          </Link>
          <br />
          <Link href="/account-recovery">
            <a>Recover your account</a>
          </Link>
        </>
      )}
    </div>
  )
}

export default withOryErrorBoundary(Login)
