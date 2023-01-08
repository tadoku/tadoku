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
