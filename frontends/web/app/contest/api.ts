import { get } from '../api'
import { Contest, rawContest } from './interfaces'

const getLatest = async (): Promise<Contest | undefined> => {
  const response = await get(`/contests/latest`)

  if (response.status !== 200) {
    return undefined
  }

  const data: rawContest = await response.json()

  return {
    id: data.id,
    start: new Date(data.start),
    end: new Date(data.end),
    open: data.open,
  }
}

const ContestApi = {
  getLatest,
}

export default ContestApi
