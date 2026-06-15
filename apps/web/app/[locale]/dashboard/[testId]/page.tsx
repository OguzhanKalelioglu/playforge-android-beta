import Link from 'next/link'
import { notFound } from 'next/navigation'
import { cookies } from 'next/headers'
import {
  ChevronLeft,
  Activity,
  Users,
  Star,
  Calendar,
  ExternalLink,
  AlertCircle,
  CheckCircle2,
  Download,
  Play,
  MessageSquare,
} from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { PackageName } from '@/components/ui/package-name'
import { ProgressBar } from '@/components/ui/progress'
import { Money } from '@/components/ui/money'
import { api, ApiError } from '@/lib/api'
import { formatDate, formatDateShort } from '@/lib/format'
import { cn } from '@/lib/utils'

export const metadata = { title: 'Test Detayı' }

const actionIcon = {
  opt_in: Users,
  download: Download,
  install: CheckCircle2,
  open: Play,
  interact: Activity,
  review: Star,
  error: AlertCircle,
}

const actionLabel: Record<string, string> = {
  opt_in: 'Testere katıldı',
  download: 'İndirme başladı',
  install: 'Yüklendi',
  open: 'Açıldı',
  interact: 'Etkileşim',
  review: 'Yorum yazıldı',
  error: 'Hata',
}

export default async function TestDetailPage({
  params,
}: {
  params: Promise<{ testId: string }>
}) {
  const { testId } = await params
  const token = (await cookies()).get('access_token')?.value ?? ''

  let test
  let activity
  let reviews
  try {
    test = await api.test(testId, token)
    activity = await api.testActivity(testId, token)
    reviews = await api.testReviews(testId, token)
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) notFound()
    return (
      <div className="container py-8">
        <div className="rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          Test yüklenemedi. Birazdan tekrar deneyin.
        </div>
      </div>
    )
  }

  const total = 25
  const stats = {
    optIn: test.assignments?.filter((a) => a.status !== 'pending').length ?? 0,
    installed: test.progress?.installed ?? 0,
    engaged: test.progress?.engaged ?? 0,
    reviews: reviews.length,
  }

  return (
    <div className="container py-8">
      <Link
        href="/dashboard"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ChevronLeft className="h-4 w-4" /> Testlerim
      </Link>

      <header className="mt-3 flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <PackageName pkg={test.package_name} />
            <StatusBadge status={test.status} />
          </div>
          <h1 className="mt-2 text-2xl font-semibold tracking-tightish">Test detayı</h1>
          {test.started_at && (
            <p className="mt-1 text-sm text-muted-foreground">
              Başladı: {formatDate(test.started_at)}
              {test.ends_at && ` · Bitiş: ${formatDate(test.ends_at)}`}
            </p>
          )}
        </div>
        {test.test_link && (
          <Button asChild variant="outline" size="sm">
            <a href={test.test_link} target="_blank" rel="noopener noreferrer">
              Test linki <ExternalLink className="ml-2 h-3 w-3" />
            </a>
          </Button>
        )}
      </header>

      {/* Progress overview */}
      <div className="mt-8 grid gap-3 sm:grid-cols-4">
        <StatCard label="Opt-in" value={stats.optIn} total={total} />
        <StatCard label="Yüklendi" value={stats.installed} total={total} />
        <StatCard label="Etkileşim" value={stats.engaged} total={total} />
        <StatCard label="Yorum" value={stats.reviews} total={10} />
      </div>

      <div className="mt-8 grid gap-6 lg:grid-cols-[1fr_360px]">
        {/* Activity timeline */}
        <section>
          <h2 className="mb-3 text-sm font-medium text-muted-foreground">Aktivite zaman çizgisi</h2>
          {activity.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center text-sm text-muted-foreground">
                Henüz aktivite yok. Test başladığında buraya düşecek.
              </CardContent>
            </Card>
          ) : (
            <ol className="relative space-y-2 border-l border-border pl-6">
              {activity.map((ev) => {
                const Icon = actionIcon[ev.action as keyof typeof actionIcon] ?? Activity
                return (
                  <li key={ev.id} className="relative">
                    <div
                      className={cn(
                        'absolute -left-9 flex h-6 w-6 items-center justify-center rounded-full border-2 border-background',
                        ev.success ? 'bg-success/15 text-success' : 'bg-destructive/15 text-destructive'
                      )}
                    >
                      <Icon className="h-3 w-3" />
                    </div>
                    <div className="rounded-md border bg-card p-3">
                      <div className="flex items-center justify-between gap-2">
                        <span className="text-sm font-medium">
                          {actionLabel[ev.action] ?? ev.action}
                        </span>
                        <span className="text-xs text-muted-foreground tabular-nums">
                          {formatDate(ev.performed_at)}
                        </span>
                      </div>
                      {!ev.success && ev.error_message && (
                        <p className="mt-1 text-xs text-destructive">{ev.error_message}</p>
                      )}
                    </div>
                  </li>
                )
              })}
            </ol>
          )}
        </section>

        {/* Sidebar */}
        <aside className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Tester durumu</CardTitle>
              <CardDescription>25 hesaptan aktif olanlar</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2 text-sm">
              {(test.assignments ?? []).slice(0, 10).map((a) => (
                <div key={a.id} className="flex items-center justify-between gap-2">
                  <span className="truncate font-mono text-xs">{a.tester_email}</span>
                  <StatusBadge status={a.status} />
                </div>
              ))}
              {(test.assignments?.length ?? 0) > 10 && (
                <p className="text-xs text-muted-foreground">
                  +{test.assignments!.length - 10} daha
                </p>
              )}
            </CardContent>
          </Card>

          {reviews.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Yorumlar</CardTitle>
                <CardDescription>{reviews.length} adet</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                {reviews.slice(0, 3).map((r) => (
                  <div key={r.id} className="rounded-md border bg-muted/30 p-3">
                    <div className="flex items-center gap-1 text-warning">
                      {Array.from({ length: r.rating }).map((_, i) => (
                        <Star key={i} className="h-3 w-3 fill-current" />
                      ))}
                    </div>
                    <p className="mt-1.5 text-xs leading-relaxed">{r.review_text}</p>
                    <p className="mt-1.5 text-xs text-muted-foreground">{formatDateShort(r.posted_at)}</p>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}
        </aside>
      </div>
    </div>
  )
}

function StatCard({ label, value, total }: { label: string; value: number; total: number }) {
  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-baseline justify-between">
          <span className="text-xs text-muted-foreground">{label}</span>
          <span className="text-xs tabular-nums text-muted-foreground">{value}/{total}</span>
        </div>
        <div className="mt-2 text-2xl font-semibold tabular-nums">{value}</div>
        <ProgressBar value={value} total={total} size="sm" className="mt-3" />
      </CardContent>
    </Card>
  )
}
