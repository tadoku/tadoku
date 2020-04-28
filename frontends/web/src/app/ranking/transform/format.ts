import { GlobalLanguage, languageNameByCode } from '../database'

export const formatScore = (amount: number) => Math.round(amount * 10) / 10

export const scoreLabel = (languageCode: string) => {
  if (languageCode == GlobalLanguage.code) {
    return 'Overall score'
  }

  return `Score for ${languageNameByCode(languageCode)}`
}

export const formatPoints = (amount: number): string => {
  switch (amount) {
    case 0:
      return 'No points'
    case 1:
      return '1 point'
    default:
      return `${formatScore(amount)} points`
  }
}
