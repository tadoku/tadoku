import { activityColors } from '@app/common/variables'

export const formatScore = (
  amount: number | undefined,
  maximumFractionDigits: number = 1,
) => {
  if (!amount) {
    return 'N/A'
  }

  return new Intl.NumberFormat('en-US', { maximumFractionDigits }).format(
    amount,
  )
}

export const formatUnit = (amount: number, unit: string) =>
  `${unit}${amount !== 1 ? 's' : ''}`

export const colorForActivity = (id: number) =>
  activityColors[id % activityColors.length]
