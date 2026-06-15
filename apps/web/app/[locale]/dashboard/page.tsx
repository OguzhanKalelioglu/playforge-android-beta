import Link from 'next/link'
import { Plus, TestTube2, Activity, Calendar, ChevronRight, AlertCircle } from 'lucide-react'
import { cookies } from 'next/headers'
import { getTranslations } from 'next-intl/server'

import { Button } from '@/components/ui/button'
import { StatusBadge } from '@/components/ui/status-badge'
import { PackageName } from '@/components/ui/package-name'
import { ProgressBar } from '@/components/ui/progress'
import { EmptyState } from '@/components/ui/empty-state'
import { api } from '@/lib/api'
import { formatDateShort } from '@/lib/format'
import { type Locale } from '@/lib/brand'

export default async function DashboardPage({ params }: { params: Promise<{ locale: Locale }> }) {
  const { locale } = await params
  const t = await getTranslations('dashboard')
  const tCommon = await getTranslations('common')
  const token = (await cookies()).get('access_token')?.value ?? ''
  let tests: Awaited<ReturnType<typeof api.tests>> = []
  let error: string | null = null

  try {
    tests = await api.tests(token)
  } catch {
    error = t('loadingError')
  }

  const active = tests.filter((t) => t.status === 'active' || t.status === 'pending')
  const completed = tests.filter((t) => t.status === 'completed' || t.status === 'failed' || t.status === 'cancelled')

  return (
    <div className="container py-8">
      <header className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tightish">{t('title')}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {active.length > 0
              ? t('subtitle', { active: active.length, completed: completed.length })
              : t('subtitleEmpty')}
          </p>
        </div>
        <Button asChild>
          <Link href={`/${locale}/dashboard/new`}>
            <Plus className="mr-2 h-4 w-4" /> {t('newTest')}
          </Link>
        </Button>
      </header>

      {error && (
        <div className="mb-6 flex items-start gap-3 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          {error}
        </div>
      )}

      {tests.length === 0 ? (
        <EmptyState
          icon={TestTube2}
          title={t('empty.title')}
          description={t('empty.description')}
          actionLabel={t('empty.action')}
          actionHref={`/${locale}/dashboard/new`}
        />
      ) : (
        <div className="space-y-6">
          {active.length > 0 && (
            <section>
              <h2 className="mb-3 text-sm font-medium text-muted-foreground">{t('activeSection')}</h2>
              <ul className="grid gap-3">
                {active.map((test) => (
                  <li key={test.id}>
                    <TestRow test={test} locale={locale} t={t} />
                  </li>
                ))}
              </ul>
            </section>
          )}

          {completed.length > 0 && (
            <section>
              <h2 className="mb-3 text-sm font-medium text-muted-foreground">{t('historySection')}</h2>
              <ul className="grid gap-3">
                {completed.map((test) => (
                  <li key={test.id}>
                    <TestRow test={test} locale={locale} t={t} />
                  </li>
                ))}
              </ul>
            </section>
          )}
        </div>
      )}
    </div>
  )
}

function TestRow({
  test,
  locale,
  t,
}: {
  test: Awaited<ReturnType<typeof api.tests>>[number]
  locale: Locale
  t: Awaited<ReturnType<typeof getTranslations<'dashboard'>>>
}) {
  const total = test.progress?.total ?? 25
  const done = test.progress?.opt_in ?? 0
  const installed = test.progress?.installed ?? 0
  const reviews = test.progress?.reviewed ?? 0

  return (
    <Link
      href={`/${locale}/dashboard/${test.id}`}
      className="block rounded-lg border bg-card p-5 transition-colors hover:border-primary/40 hover:bg-card/80"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1 space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <PackageName pkg={test.package_name} />
            <StatusBadge status={test.status} />
          </div>
          <ProgressBar
            value={installed}
            total={total}
            showLabel
            tone={test.status === 'failed' ? 'warning' : 'primary'}
            size="md"
          />
          <dl className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
            <div className="flex items-center gap-1.5">
              <Activity className="h-3 w-3" />
              <span>{t('optIn', { done, total })}</span>
            </div>
            <div className="flex items-center gap-1.5">
              <TestTube2 className="h-3 w-3" />
              <span>{t('reviews', { count: reviews })}</span>
            </div>
            <div className="flex items-center gap-1.5">
              <Calendar className="h-3 w-3" />
              <span>
                {test.ends_at
                  ? t('endsAt', { date: formatDateShort(test.ends_at) })
                  : t('startsAt', { date: formatDateShort(test.created_at) })}
              </span>
            </div>
          </dl>
        </div>
        <ChevronRight className="mt-1 h-4 w-4 shrink-0 text-muted-foreground" />
      </div>
    </Link>
  )
}
