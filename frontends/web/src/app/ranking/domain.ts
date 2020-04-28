import { languageByCode } from './database'
import { Contest } from '../contest/interfaces'

export const validateAmount = (amount: string): boolean =>
  Number(amount) !== NaN && Number(amount) > 0

export const validateLanguageCode = (code: string): boolean =>
  code != '' && languageByCode[code] !== undefined

export const isContestActive = (contest: Contest): boolean =>
  contest.open && contest.end > new Date()
