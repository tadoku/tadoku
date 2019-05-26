import { get } from '../api'
import { Contest, rawContest } from './interfaces'

const getContest = async (contestId: number): Promise<Contest | undefined> => {
  const response = await get(`/contests/${contestId}`)

  if (response.status !== 200) {
    return undefined
  }

  const data: rawContest = await response.json()

  return {
    id: data.id,
    description: data.description,
    start: new Date(data.start),
    end: new Date(data.end),
    open: data.open,
  }
}

const getLatest = async (): Promise<Contest | undefined> => {
  const response = await get(`/contests/latest`)

  if (response.status !== 200) {
    return undefined
  }

  const data: rawContest = await response.json()

  return {
    id: data.id,
    description: data.description,
    start: new Date(data.start),
    end: new Date(data.end),
    open: data.open,
  }
}

const ContestApi = {
  get: getContest,
  getLatest,
}

export default ContestApi
