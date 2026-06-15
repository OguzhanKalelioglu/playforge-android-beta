import { CheckCircle2, Clock, XCircle, AlertCircle, Pause, Loader2 } from 'lucide-react'
import { Badge } from './badge'
import { cn } from '@/lib/utils'

type Status =
  | 'active'
  | 'pending'
  | 'completed'
  | 'failed'
  | 'cancelled'
  | 'in_progress'
  | 'paused'
  | 'warming'
  | 'cooling'
  | 'disabled'

const map: Record<
  Status,
  { label: string; variant: 'success' | 'warning' | 'destructive' | 'info' | 'muted'; icon: React.ComponentType<{ className?: string }> }
> = {
  active: { label: 'Aktif', variant: 'success', icon: CheckCircle2 },
  in_progress: { label: 'Devam Ediyor', variant: 'info', icon: Loader2 },
  pending: { label: 'Beklemede', variant: 'warning', icon: Clock },
  completed: { label: 'Tamamlandı', variant: 'success', icon: CheckCircle2 },
  failed: { label: 'Başarısız', variant: 'destructive', icon: XCircle },
  cancelled: { label: 'İptal Edildi', variant: 'muted', icon: Pause },
  paused: { label: 'Duraklatıldı', variant: 'muted', icon: Pause },
  warming: { label: 'Hazırlanıyor', variant: 'warning', icon: Loader2 },
  cooling: { label: 'Soğuma', variant: 'muted', icon: AlertCircle },
  disabled: { label: 'Devre Dışı', variant: 'muted', icon: AlertCircle },
}

export function StatusBadge({ status, className }: { status: string; className?: string }) {
  const config = map[status as Status] ?? { label: status, variant: 'muted' as const, icon: AlertCircle }
  const Icon = config.icon
  return (
    <Badge variant={config.variant} className={cn('font-medium', className)}>
      <Icon className="h-3 w-3" />
      <span>{config.label}</span>
    </Badge>
  )
}
