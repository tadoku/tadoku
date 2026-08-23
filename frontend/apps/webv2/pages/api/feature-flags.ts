import type { NextApiRequest, NextApiResponse } from 'next'
import { sdkServer as ory } from '@app/common/ory'
import { decisionsForSession } from '@app/feature-flags/server'
import {
  defaultFeatureFlagDecisions,
} from '@app/feature-flags/registry'
import type { FeatureFlagResponse } from '@app/feature-flags/registry'

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse<FeatureFlagResponse | { error: string }>,
) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', 'GET')
    res.status(405).json({ error: 'method not allowed' })
    return
  }

  res.setHeader('Cache-Control', 'private, no-store')

  try {
    const cookie = req.headers.cookie
    if (!cookie) {
      res.status(200).json({ decisions: { ...defaultFeatureFlagDecisions } })
      return
    }

    const { data: session } = await ory.toSession(undefined, cookie)
    const decisions = await decisionsForSession(session)
    res.status(200).json({ decisions })
  } catch {
    res.status(200).json({ decisions: { ...defaultFeatureFlagDecisions } })
  }
}
