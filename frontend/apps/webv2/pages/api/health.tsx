import type { NextApiRequest, NextApiResponse } from 'next'

export default function handler(
  _: NextApiRequest,
  res: NextApiResponse<unknown>,
) {
  res.status(200).json({ status: 'ok' })
}
