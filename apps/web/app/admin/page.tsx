import { cookies } from 'next/headers'
import {
  Users,
  TestTube2,
  Activity,
  Smartphone,
  TrendingUp,
  AlertCircle,
  CheckCircle2,
  type LucideIcon,
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Money } from '@/components/ui/money'

const API_BASE = process.env.INTERNAL_API_URL ?? 'http://127.0.0.1:8080'

interface AdminMetrics {
  active_tests: number
  pending_tests: number
  total_testers: number
  active_testers: number
  warming_testers: number
  emulators_ready: number
  emulators_total: number
  revenue_month_try: number
  failed_tasks_24h: number
}

async function getMetrics(token: string): Promise<AdminMetrics | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/admin/metrics`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
    })
    if (!res.ok) return null
    return res.json()
  } catch {
    return null
  }
}

export const metadata = { title: 'Admin — Genel Bakış' }

export default async function AdminOverviewPage() {
  const token = (await cookies()).get('access_token')?.value ?? ''
  const m = await getMetrics(token)

  return (
    <div className="container py-8">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tightish">Genel bakış</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Operasyonel metrikler ve sistem sağlığı.
        </p>
      </header>

      {/* Top metrics */}
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          icon={TestTube2}
          label="Aktif testler"
          value={m?.active_tests ?? 0}
          hint={`${m?.pending_tests ?? 0} beklemede`}
          tone="primary"
        />
        <MetricCard
          icon={Users}
          label="Tester havuzu"
          value={m?.total_testers ?? 0}
          hint={`${m?.active_testers ?? 0} aktif, ${m?.warming_testers ?? 0} warming`}
          tone="info"
        />
        <MetricCard
          icon={Smartphone}
          label="Emulator durumu"
          value={`${m?.emulators_ready ?? 0}/${m?.emulators_total ?? 25}`}
          hint={m?.emulators_ready === m?.emulators_total ? 'Tümü hazır' : 'Bazıları çevrimdışı'}
          tone={(m?.emulators_ready ?? 0) === (m?.emulators_total ?? 25) ? 'success' : 'warning'}
        />
        <MetricCard
          icon={TrendingUp}
          label="Bu ay gelir"
          value={<Money amount={m?.revenue_month_try ?? 0} short />}
          hint="TRY"
          tone="primary"
        />
      </div>

      {/* Health section */}
      <div className="mt-8 grid gap-3 sm:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Son 24 saat</CardDescription>
            <CardTitle className="text-base">Hata oranı</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-baseline gap-2">
              <span className="text-3xl font-semibold tabular-nums">{m?.failed_tasks_24h ?? 0}</span>
              <span className="text-sm text-muted-foreground">başarısız task</span>
            </div>
            <Badge
              variant={(m?.failed_tasks_24h ?? 0) > 5 ? 'destructive' : 'success'}
              className="mt-3"
            >
              {(m?.failed_tasks_24h ?? 0) > 5 ? (
                <>
                  <AlertCircle className="h-3 w-3" /> İncelenmeli
                </>
              ) : (
                <>
                  <CheckCircle2 className="h-3 w-3" /> Sağlıklı
                </>
              )}
            </Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Emulator farm</CardDescription>
            <CardTitle className="text-base">Sistem sağlığı</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm">
              <li className="flex items-center justify-between">
                <span className="text-muted-foreground">ADB bağlantısı</span>
                <Badge variant="success">OK</Badge>
              </li>
              <li className="flex items-center justify-between">
                <span className="text-muted-foreground">Appium server</span>
                <Badge variant="success">OK</Badge>
              </li>
              <li className="flex items-center justify-between">
                <span className="text-muted-foreground">Orchestrator</span>
                <Badge variant="success">OK</Badge>
              </li>
              <li className="flex items-center justify-between">
                <span className="text-muted-foreground">PostgreSQL</span>
                <Badge variant="success">OK</Badge>
              </li>
            </ul>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Queue</CardDescription>
            <CardTitle className="text-base">Asynq</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm">
              <li className="flex items-center justify-between">
                <span className="text-muted-foreground">Bekleyen</span>
                <span className="font-mono tabular-nums">—</span>
              </li>
              <li className="flex items-center justify-between">
                <span className="text-muted-foreground">Aktif worker</span>
                <span className="font-mono tabular-nums">10</span>
              </li>
              <li className="flex items-center justify-between">
                <span className="text-muted-foreground">Retry</span>
                <span className="font-mono tabular-nums">—</span>
              </li>
              <li className="flex items-center justify-between">
                <span className="text-muted-foreground">Dead (24h)</span>
                <span className="font-mono tabular-nums">—</span>
              </li>
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function MetricCard({
  icon: Icon,
  label,
  value,
  hint,
  tone,
}: {
  icon: LucideIcon
  label: string
  value: React.ReactNode
  hint?: string
  tone: 'primary' | 'success' | 'warning' | 'info'
}) {
  const toneClass = {
    primary: 'text-primary',
    success: 'text-success',
    warning: 'text-warning',
    info: 'text-info',
  }[tone]

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-start justify-between">
          <span className="text-xs text-muted-foreground">{label}</span>
          <Icon className={`h-4 w-4 ${toneClass}`} strokeWidth={1.75} />
        </div>
        <div className="mt-2 text-2xl font-semibold tabular-nums">{value}</div>
        {hint && <p className="mt-1 text-xs text-muted-foreground">{hint}</p>}
      </CardContent>
    </Card>
  )
}
