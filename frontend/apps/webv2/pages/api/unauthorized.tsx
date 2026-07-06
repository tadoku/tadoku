import type { NextApiRequest, NextApiResponse } from 'next'
import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()

export default function handler(
  _: NextApiRequest,
  res: NextApiResponse<unknown>,
) {
  const cookieAttributes = [
    'ory_kratos_session=deleted',
    'Path=/',
    `Domain=${publicRuntimeConfig.cookieDomain}`,
    'Max-Age=0',
    'Expires=Thu, 01 Jan 1970 00:00:00 GMT',
    'SameSite=Lax',
  ]
  if (publicRuntimeConfig.cookieSecure) {
    cookieAttributes.push('Secure')
  }

  res.setHeader(
    'Set-Cookie',
    cookieAttributes.join('; '),
  )
  res.status(200).redirect(publicRuntimeConfig.authUiUrl)
}
