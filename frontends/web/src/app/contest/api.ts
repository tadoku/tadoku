import { formatRFC3339 } from 'date-fns'
import { get, put, post } from '../api'
import { Contest, RawContest } from './interfaces'
import { contestMapper } from './transform'

const getContest = async (contestId: number): Promise<Contest | undefined> => {
  const response = await get(`/contests/${contestId}`)

  if (response.status !== 200) {
    return undefined
  }

  const data: RawContest = await response.json()

  return contestMapper.fromRaw(data)
}

const getAll = async (limit?: number): Promise<Contest[]> => {
  let queryString = limit ? `?limit=${limit}` : ''
  const response = await get(`/contests${queryString}`)

  if (response.status !== 200) {
    return []
  }

  const data: RawContest[] = await response.json()

  return data.map(contestMapper.fromRaw)
}

const create = async (payload: {
  start: Date
  end: Date
  description: string
  open: boolean
}): Promise<boolean> => {
  const response = await post(`/contests`, {
    body: {
      start: formatRFC3339(payload.start),
      end: formatRFC3339(payload.end),
      description: payload.description,
      open: payload.open,
    },
  })

  return response.status === 201
}

const update = async (
  id: number,
  payload: {
    start: Date
    end: Date
    description: string
    open: boolean
  },
): Promise<boolean> => {
  const response = await put(`/contests/${id}`, {
    body: {
      start: payload.start.toISOString(),
      end: payload.end.toISOString(),
      description: payload.description,
      open: payload.open,
    },
  })

  return response.status === 204
}

const ContestApi = {
  get: getContest,
  getAll,
  create,
  update,
}

export default ContestApi
