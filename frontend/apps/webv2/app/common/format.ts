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

export const formatUnit = (
  amount: number | undefined | null,
  unit: string | undefined | null,
) => {
  if (!unit) return ''
  return `${unit.toLowerCase()}${amount !== 1 ? 's' : ''}`
}

export const colorForActivity = (id: number) =>
  activityColors[id % activityColors.length]

export function formatArray<T>(elements: T[], format: (element: T) => any) {
  if (elements.length === 0) {
    return ''
  }
  if (elements.length === 1) {
    return format(elements[0])
  }
  if (elements.length === 2) {
    return [format(elements[0]), ' and ', format(elements[1])]
  }

  return elements
    .map(it => format(it))
    .map((node, i) => [
      i > 0 && i != elements.length - 1 ? ', ' : null,
      i === elements.length - 1 ? ', and ' : null,
      node,
    ])
    .filter(it => it != null)
}
