import type { NextApiRequest, NextApiResponse } from 'next'
import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()

export default function handler(
  _: NextApiRequest,
  res: NextApiResponse<unknown>,
) {
  res.setHeader(
    'Set-Cookie',
    `ory_kratos_session=deleted; path=/; domain=${publicRuntimeConfig.cookieDomain}; expires=Thu, 01 Jan 1970 00:00:00 GMT`,
  )
  res.status(200).redirect(publicRuntimeConfig.authUiUrl)
}
