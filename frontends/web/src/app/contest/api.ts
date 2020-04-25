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

const getAll = async (limit?: number): Promise<Contest[]> => {
  let queryString = limit ? `?limit=${limit}` : ''
  const response = await get(`/contests${queryString}`)

  if (response.status !== 200) {
    return []
  }

  const data: RawContest[] = await response.json()

  return data.map(ContestMapper.fromRaw)
}

const ContestApi = {
  get: getContest,
  getAll,
}

export default ContestApi
