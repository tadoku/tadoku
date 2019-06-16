import { get } from '../api'
import { Contest, rawContest } from './interfaces'
import { RawToContestMapper } from './transform'

const getContest = async (contestId: number): Promise<Contest | undefined> => {
  const response = await get(`/contests/${contestId}`)

  if (response.status !== 200) {
    return undefined
  }

  const data: rawContest = await response.json()

  return RawToContestMapper(data)
}

const getLatest = async (): Promise<Contest | undefined> => {
  const response = await get(`/contests/latest`)

  if (response.status !== 200) {
    return undefined
  }

  const data: rawContest = await response.json()

  return RawToContestMapper(data)
}

const ContestApi = {
  get: getContest,
  getLatest,
}

export default ContestApi
