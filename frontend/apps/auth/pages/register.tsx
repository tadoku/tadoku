import {
  SelfServiceRegistrationFlow,
  SubmitSelfServiceRegistrationFlowBody,
} from '@ory/client'
import type { NextPage } from 'next'
import { useEffect, useState } from 'react'
import Flow from '../src/ui/Flow'
import ory from '../src/ory'
import { useSession } from '../src/session'
import { useRouter } from 'next/router'
import { handleFlowError } from '../src/errors'
import { AxiosError } from 'axios'
import Link from 'next/link'

interface Props {}

const Register: NextPage<Props> = () => {
  const [flow, setFlow] = useState<SelfServiceRegistrationFlow>()
  const [session, setSession] = useSession()
  const router = useRouter()
  const { flow: flowId, return_to: returnTo } = router.query

  // In this effect we either initiate a new registration flow, or we fetch an existing registration flow.
  useEffect(() => {
    // If the router is not ready yet, or we already have a flow, do nothing.
    if (!router.isReady || flow) {
      return
    }

    if (session) {
      router.replace('/')
      return
    }

    // If ?flow=.. was in the URL, we fetch it
    if (flowId) {
      ory
        .getSelfServiceRegistrationFlow(String(flowId))
        .then(({ data }) => {
          // We received the flow - let's use its data and render the form!
          setFlow(data)
        })
        .catch(handleFlowError(router, 'registration', setFlow))
      return
    }

    // Otherwise we initialize it
    ory
      .initializeSelfServiceRegistrationFlowForBrowsers(
        returnTo ? String(returnTo) : undefined,
      )
      .then(({ data }) => {
        setFlow(data)
      })
      .catch(handleFlowError(router, 'registration', setFlow))
  }, [flowId, router, router.isReady, returnTo, flow, session])

  const onSubmit = async (data: SubmitSelfServiceRegistrationFlowBody) => {
    await router.push(`/register?flow=${flow?.id}`, undefined, {
      shallow: true,
    })

    ory
      .submitSelfServiceRegistrationFlow(String(flow?.id), data)
      .then(async ({ data }) => {
        const res = await ory.toSession()
        setSession(res.data)

        return router.push(flow?.return_to || '/')
      })
      .catch(handleFlowError(router, 'registration', setFlow))
      .catch(async (err: AxiosError) => {
        // If the previous handler did not catch the error it's most likely a form validation error
        if (err.response?.status === 400) {
          // Yup, it is!
          setFlow(err.response?.data as SelfServiceRegistrationFlow | undefined)
          return
        }

        return Promise.reject(err)
      })
  }

  return (
    <div>
      <h1 className="title mb-4">Create account</h1>
      <div className="card">
        <Flow flow={flow} method="password" onSubmit={onSubmit} />
      </div>

      <div className="h-stack items-center space-x-2 mt-4 justify-center text-xs">
        <span className="text-gray-500">Already have an account?</span>
        <Link href="/login" className="btn ghost small">
          Log in now
        </Link>
      </div>
    </div>
  )
}

export default Register
