import type { NextApiRequest, NextApiResponse } from 'next'
import { appUrlsForHost } from 'ui/app-urls'

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  const urls = appUrlsForHost(req.headers.host)
  const returnTo = req.headers.referer || urls.adminUrl
  const loginUrl = `${
    urls.authUiUrl
  }/login?return_to=${encodeURIComponent(returnTo)}`
  res.redirect(302, loginUrl)
}
