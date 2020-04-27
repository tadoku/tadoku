import { ContestLog, RawContestLog } from './../interfaces'
import { Mapper } from '../../interfaces'
import { Serializer } from '../../cache'

const rawToContestLogMapper: Mapper<RawContestLog, ContestLog> = raw => ({
  id: raw.id,
  contestId: raw.contest_id,
  userId: raw.user_id,
  languageCode: raw.language_code,
  mediumId: raw.medium_id,
  amount: raw.amount,
  adjustedAmount: raw.adjusted_amount,
  description: raw.description ? raw.description : undefined,
  date: new Date(raw.date),
})

const contestLogToRawMapper: Mapper<
  ContestLog,
  RawContestLog
> = contestLog => ({
  id: contestLog.id,
  contest_id: contestLog.contestId,
  user_id: contestLog.userId,
  language_code: contestLog.languageCode,
  medium_id: contestLog.mediumId,
  amount: contestLog.amount,
  adjusted_amount: contestLog.adjustedAmount,
  description: contestLog.description ? contestLog.description : '',
  date: contestLog.date.toISOString(),
})

export const contestLogsSerializer: Serializer<ContestLog[]> = {
  serialize: data => {
    const raw = data.map(contestLogToRawMapper)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return raw.map(rawToContestLogMapper)
  },
}

export const contestLogMapper = {
  toRaw: contestLogToRawMapper,
  fromRaw: rawToContestLogMapper,
}
