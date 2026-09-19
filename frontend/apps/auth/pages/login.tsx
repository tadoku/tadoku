import { LoginFlow, UpdateLoginFlowBody } from '@ory/kratos-client'
import type { NextPage } from 'next'
import { useEffect, useState } from 'react'
import Flow from '../src/ui/Flow'
import ory from '../src/ory'
import { AxiosError } from 'axios'
import { useLogoutHandler, useSession } from '../src/session'
import { useRouter } from 'next/router'
import { handleFlowError } from '../src/errors'
import Link from 'next/link'
import { ErrorFallback, withOryErrorBoundary } from '../src/OryErrorBoundary'

interface Props {}

const Login: NextPage<Props> = () => {
  const [flow, setFlow] = useState<LoginFlow>()
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

    if (session && !flowId && refresh !== 'true' && !aal) {
      router.replace('/')
      return
    }

    // If ?flow=.. was in the URL, we fetch it
    if (flowId) {
      ory
        .getLoginFlow({ id: String(flowId) })
        .then(({ data }) => {
          setFlow(data)
        })
        .catch(handleFlowError(router, 'login', setFlow))
      return
    }

    ory
      .createBrowserLoginFlow({
        refresh: refresh === 'true',
        aal: aal ? String(aal) : undefined,
        returnTo: returnTo ? String(returnTo) : undefined,
      })
      .then(({ data }) => {
        setFlow(data)
      })
      .catch(handleFlowError(router, 'login', setFlow))
      .catch(err => setError(err))
  }, [
    flowId,
    router,
    router.isReady,
    aal,
    refresh,
    returnTo,
    flow,
    error,
    session,
  ])

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

  const onSubmit = async (data: UpdateLoginFlowBody) => {
    await router.push(`/login?flow=${flow?.id}`, undefined, {
      shallow: true,
    })

    return ory
      .updateLoginFlow({ flow: flow.id, updateLoginFlowBody: data })
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
          setFlow(err.response?.data as LoginFlow | undefined)
          return
        }

        return Promise.reject(err)
      })
  }

  return (
    <div>
      <h1 className="title mb-4">
        {(() => {
          if (flow?.refresh) {
            return 'Confirm action'
          } else if (flow?.requested_aal === 'aal2') {
            return 'Two-factor authentication'
          }
          return 'Log in'
        })()}
      </h1>

      <div className="card relative">
        <Flow flow={flow} onSubmit={onSubmit} />

        <Link
          href="/account-recovery"
          className="btn ghost absolute right-7 bottom-7 text-sm small"
        >
          Forgot password?
        </Link>
      </div>

      {aal || refresh ? (
        <a onClick={onLogout} className="btn ghost">
          Log out
        </a>
      ) : (
        <div className="h-stack justify-center mt-4 items-center space-x-2 text-xs">
          <span className="text-gray-500">Don&apos;t have an account?</span>
          <Link href="/register" className="btn ghost small">
            Sign up now
          </Link>
        </div>
      )}
    </div>
  )
}

export default withOryErrorBoundary(Login)
