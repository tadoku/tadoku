import {
  VerificationFlow,
  UpdateVerificationFlowBody,
} from '@ory/kratos-client'
import type { NextPage } from 'next'
import { useEffect, useState } from 'react'
import Flow from '../src/ui/Flow'
import ory from '../src/ory'
import { useRouter } from 'next/router'
import { handleFlowError } from '../src/errors'
import { AxiosError } from 'axios'

interface Props {}

const Verification: NextPage<Props> = () => {
  const [flow, setFlow] = useState<VerificationFlow>()
  const router = useRouter()
  const { flow: flowId, return_to: returnTo } = router.query

  // In this effect we either initiate a new registration flow, or we fetch an existing registration flow.
  useEffect(() => {
    // If the router is not ready yet, or we already have a flow, do nothing.
    if (!router.isReady || flow) {
      return
    }

    // If ?flow=.. was in the URL, we fetch it
    if (flowId) {
      ory
        .getVerificationFlow({ id: String(flowId) })
        .then(({ data }) => {
          // We received the flow - let's use its data and render the form!
          setFlow(data)
        })
        .catch((err: AxiosError) => {
          switch (err.response?.status) {
            case 410:
            // Status code 410 means the request has expired - so let's load a fresh flow!
            case 403:
              // Status code 403 implies some other issue (e.g. CSRF) - let's reload!
              return router.push('/verification')
          }

          throw err
        })
      return
    }

    // Otherwise we initialize it
    ory
      .createBrowserVerificationFlow({
        returnTo: returnTo ? String(returnTo) : undefined,
      })
      .then(({ data }) => {
        setFlow(data)
      })
      .catch((err: AxiosError) => {
        switch (err.response?.status) {
          case 400:
            // Status code 400 implies the user is already signed in
            return router.push('/')
        }

        throw err
      })
  }, [flowId, router, router.isReady, returnTo, flow])

  const onSubmit = async (data: UpdateVerificationFlowBody) => {
    await router.push(`/verification?flow=${flow?.id}`, undefined, {
      shallow: true,
    })

    return ory
      .updateVerificationFlow({
        flow: String(flow?.id),
        updateVerificationFlowBody: data,
      })
      .then(async ({ data }) => {
        setFlow(data)
      })
      .catch(handleFlowError(router, 'verification', setFlow))
      .catch(async (err: AxiosError) => {
        // If the previous handler did not catch the error it's most likely a form validation error
        if (err.response?.status === 400) {
          // Yup, it is!
          setFlow(err.response?.data as VerificationFlow | undefined)
          return
        }

        return Promise.reject(err)
      })
  }

  return (
    <div>
      <h1 className="title mb-4">Account verification</h1>
      <Flow flow={flow} onSubmit={onSubmit} />
    </div>
  )
}

export default Verification
