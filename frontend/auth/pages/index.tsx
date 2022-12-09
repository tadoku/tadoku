import { SelfServiceSettingsFlow } from '@ory/client'
import type { NextPage } from 'next'
import { useEffect, useState } from 'react'
import Flow from '../src/ui/Flow'
import ory from '../src/ory'
import { AxiosError } from 'axios'
import { useSession } from '../src/session'
import { useRouter } from 'next/router'
import { handleFlowError } from '../src/errors'
import MessagesList from '../src/ui/MessagesList'

interface Props {}

const Settings: NextPage<Props> = () => {
  const [flow, setFlow] = useState(
    undefined as SelfServiceSettingsFlow | undefined,
  )
  const [session, setSession] = useSession()
  const router = useRouter()
  const { flow: flowId, return_to: returnTo } = router.query

  useEffect(() => {
    // Skip if we aren't ready
    if (!router.isReady || flow) {
      return
    }

    if (!session) {
      router.replace('/login')
      return
    }

    // If ?flow=.. was in the URL, we fetch it
    if (flowId) {
      ory
        .getSelfServiceSettingsFlow(String(flowId))
        .then(({ data }) => {
          setFlow(data)
        })
        .catch(handleFlowError(router, 'settings', setFlow))
      return
    }

    ory
      .initializeSelfServiceSettingsFlowForBrowsers(
        returnTo ? String(returnTo) : undefined,
      )
      .then(({ data }) => {
        console.log(data)
        setFlow(data)
      })
      .catch(handleFlowError(router, 'settings', setFlow))
  }, [flowId, router, router.isReady, returnTo, flow])

  if (!flow) {
    return null
  }

  const onSubmit = async (data: any) => {
    if (flow === undefined) {
      console.error('no settings flow available to use')
      return
    }

    await router.push(`/settings?flow=${flow?.id}`, undefined, {
      shallow: true,
    })

    ory
      .submitSelfServiceSettingsFlow(flow.id, data)
      .then(async ({ data }) => {
        console.log('Submitted settings flow', data)
        setFlow(data)

        // Update session with new data
        const session = await ory.toSession()
        setSession(session.data)
      })
      .catch(handleFlowError(router, 'settings', setFlow))
      .catch(async (err: AxiosError) => {
        // If the previous handler did not catch the error it's most likely a form validation error
        if (err.response?.status === 400) {
          setFlow(err.response.data as SelfServiceSettingsFlow)
          return
        }

        debugger

        return Promise.reject(err)
      })
  }

  return (
    <div>
      <h1 className="title">Settings</h1>
      <MessagesList messages={flow?.ui.messages} />
      <div className="v-stack md:h-stack md:space-y-0 w-full mt-4">
        <div className="flex-grow card">
          <h2 className="subtitle mb-4">Update profile</h2>
          <Flow
            flow={flow}
            method="profile"
            onSubmit={onSubmit}
            hideGlobalMessages
          />
        </div>
        <div className="flex-grow card">
          <h2 className="subtitle mb-4">Change password</h2>
          <Flow
            flow={flow}
            method="password"
            onSubmit={onSubmit}
            hideGlobalMessages
          />
        </div>
      </div>
    </div>
  )
}

export default Settings
