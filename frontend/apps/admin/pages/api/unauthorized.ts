import type { NextApiRequest, NextApiResponse } from 'next'
import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  const returnTo = req.headers.referer || `https://admin.tadoku.app`
  const loginUrl = `${publicRuntimeConfig.authUiUrl}/login?return_to=${encodeURIComponent(returnTo)}`
  res.redirect(302, loginUrl)
}
