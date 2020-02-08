import { get } from '../api'
import { Contest, RawContest } from './interfaces'
import { ContestMapper } from './transform'

const getContest = async (contestId: number): Promise<Contest | undefined> => {
  const response = await get(`/contests/${contestId}`)

  if (response.status !== 200) {
    return undefined
  }

  const data: RawContest = await response.json()

  return ContestMapper.fromRaw(data)
}

const getLatest = async (): Promise<Contest | undefined> => {
  const response = await get(`/contests/latest`)

  if (response.status !== 200) {
    return undefined
  }

  const data: RawContest = await response.json()

  return ContestMapper.fromRaw(data)
}

const ContestApi = {
  get: getContest,
  getLatest,
}

export default ContestApi
