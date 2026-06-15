import { cn } from '@/lib/utils'

interface ProgressBarProps {
  value: number
  total: number
  className?: string
  showLabel?: boolean
  size?: 'sm' | 'md' | 'lg'
  tone?: 'primary' | 'success' | 'warning'
}

export function ProgressBar({
  value,
  total,
  className,
  showLabel = false,
  size = 'md',
  tone = 'primary',
}: ProgressBarProps) {
  const pct = total === 0 ? 0 : Math.min(100, Math.round((value / total) * 100))
  const heightClass = size === 'sm' ? 'h-1' : size === 'lg' ? 'h-2.5' : 'h-2'
  const toneClass =
    tone === 'success' ? 'bg-success' : tone === 'warning' ? 'bg-warning' : 'bg-primary'
  return (
    <div className={cn('flex items-center gap-3', className)}>
      <div
        className={cn('flex-1 overflow-hidden rounded-full bg-muted', heightClass)}
        role="progressbar"
        aria-valuenow={pct}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-label={`İlerleme: ${value} / ${total}`}
      >
        <div
          className={cn('h-full rounded-full transition-all duration-300', toneClass)}
          style={{ width: `${pct}%` }}
        />
      </div>
      {showLabel && (
        <span className="text-xs font-medium tabular-nums text-muted-foreground">
          {value}/{total} · {pct}%
        </span>
      )}
    </div>
  )
}
