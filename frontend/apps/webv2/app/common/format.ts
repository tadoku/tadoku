export const formatScore = (
  amount: number,
  maximumFractionDigits: number = 1,
) => {
  return new Intl.NumberFormat('en-US', { maximumFractionDigits }).format(
    amount,
  )
}
