import { cn } from '@/lib/utils'
import { formatCurrency } from '@/lib/format'

interface MoneyProps {
  amount: number
  currency?: string
  className?: string
  short?: boolean
}

export function Money({ amount, currency = 'TRY', className, short = false }: MoneyProps) {
  return (
    <span className={cn('font-medium tabular-nums', className)}>
      {short
        ? `${(amount / 1000).toFixed(amount >= 10000 ? 0 : 1).replace('.0', '')}K ₺`
        : formatCurrency(amount, currency)}
    </span>
  )
}
