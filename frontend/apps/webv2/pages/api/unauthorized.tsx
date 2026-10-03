import type { NextApiRequest, NextApiResponse } from 'next'
import getConfig from 'next/config'
import { appUrlsForHost } from 'ui/app-urls'

const { publicRuntimeConfig } = getConfig()

export default function handler(
  req: NextApiRequest,
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

  res.setHeader('Set-Cookie', cookieAttributes.join('; '))
  res.status(200).redirect(appUrlsForHost(req.headers.host).authUiUrl)
}
