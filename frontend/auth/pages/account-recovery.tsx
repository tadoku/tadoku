import { SelfServiceRecoveryFlow } from '@ory/client'
import type { NextPage, GetServerSideProps } from 'next'
import { useState } from 'react'
import Flow from '../ui/Flow'
import ory from '../src/ory'
import axios from 'axios'
import { useAnonymouseRoute, useSession } from '../src/session'

interface Props {
  initialFlow: SelfServiceRecoveryFlow
}

// TODO: Refactor to client-side

const AccountRecovery: NextPage<Props> = ({ initialFlow }) => {
  const [flow, setFlow] = useState(initialFlow)

  // Would be better to do this at the layout level once the feature is available
  useAnonymouseRoute()

  const onSubmit = async (data: any) => {
    if (flow === undefined) {
      console.error('no account recovery flow available to use')
      return
    }

    try {
      console.log(data)
      const res = await ory.submitSelfServiceRecoveryFlow(flow.id, data)
    } catch (err) {
      if (
        axios.isAxiosError(err) &&
        err.response?.data &&
        err.response.status === 400
      ) {
        // TODO: figure out types
        setFlow(err.response.data as SelfServiceRecoveryFlow)
      }
    }
  }

  return (
    <div>
      <h1>Account recovery</h1>
      <Flow flow={flow} method="link" onSubmit={onSubmit} />
    </div>
  )
}

export const getServerSideProps: GetServerSideProps = async ctx => {
  try {
    const { data: initialFlow, headers } =
      await ory.initializeSelfServiceRecoveryFlowForBrowsers()

    // Proxy cookies
    if (headers['set-cookie']) {
      ctx.res.setHeader('set-cookie', headers['set-cookie'])
    }

    return { props: { initialFlow } }
  } catch (err) {
    console.error(err)
    return {
      redirect: {
        destination: '/error',
      },
      props: {},
    }
  }
}

export default AccountRecovery
