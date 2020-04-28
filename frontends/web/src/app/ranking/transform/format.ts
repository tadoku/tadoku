import { GlobalLanguage, languageNameByCode } from '../database'

export const formatScore = (amount: number) => Math.round(amount * 10) / 10

export const scoreLabel = (languageCode: string) => {
  if (languageCode == GlobalLanguage.code) {
    return 'Overall score'
  }

  return `Score for ${languageNameByCode(languageCode)}`
}

export const amountToString = (amount: number): string => {
  switch (amount) {
    case 0:
      return 'No pages'
    case 1:
      return '1 page'
    default:
      return `${formatScore(amount)} pages`
  }
}
