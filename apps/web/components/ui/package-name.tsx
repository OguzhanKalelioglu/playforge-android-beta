import { cn } from '@/lib/utils'
import { formatPackageName } from '@/lib/format'

interface PackageNameProps {
  pkg: string
  max?: number
  className?: string
}

export function PackageName({ pkg, max = 28, className }: PackageNameProps) {
  return (
    <code
      className={cn(
        'inline-block max-w-full truncate rounded bg-muted px-1.5 py-0.5 font-mono text-xs',
        className
      )}
      title={pkg}
      dir="rtl"
    >
      {formatPackageName(pkg, max)}
    </code>
  )
}
