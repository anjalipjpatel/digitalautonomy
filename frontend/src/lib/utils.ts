export { cn } from "cn"

export function formatAmount(amount: string | number): string {
  const numericAmount = Number(amount)
  return Number.isFinite(numericAmount) ? numericAmount.toFixed(2) : '0.00'
}
