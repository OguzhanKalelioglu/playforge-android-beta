'use client'

import { useState, Suspense } from 'react'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useTranslations } from 'next-intl'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { type Locale } from '@/lib/brand'

function LoginForm({ locale }: { locale: Locale }) {
  const t = useTranslations('auth')
  const router = useRouter()
  const search = useSearchParams()
  const next = search.get('next') ?? `/${locale}/dashboard`
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const schema = z.object({
    email: z.string().email(t('emailInvalid')),
    password: z.string().min(8, t('passwordTooShort')),
  })
  type FormValues = z.infer<typeof schema>

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) })

  const onSubmit = async (values: FormValues) => {
    setSubmitting(true)
    setError(null)
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        setError(body.error ?? t('loginError'))
        return
      }
      router.push(next)
    } catch {
      setError(t('loginError'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('loginTitle')}</CardTitle>
        <CardDescription>{t('loginSubtitle')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="email">{t('email')}</Label>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              placeholder={t('emailPlaceholder')}
              aria-invalid={!!errors.email}
              {...register('email')}
            />
            {errors.email && <p className="text-xs text-destructive">{errors.email.message}</p>}
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="password">{t('password')}</Label>
              <Link
                href={`/${locale}/forgot-password`}
                className="text-xs text-muted-foreground hover:text-foreground"
              >
                {t('forgotPassword')}
              </Link>
            </div>
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              aria-invalid={!!errors.password}
              placeholder={t('passwordPlaceholder')}
              {...register('password')}
            />
            {errors.password && (
              <p className="text-xs text-destructive">{errors.password.message}</p>
            )}
          </div>

          {error && (
            <div role="alert" className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          )}

          <Button type="submit" className="w-full" disabled={submitting}>
            {submitting ? t('signingIn') : t('login')}
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          {t('noAccount')}{' '}
          <Link href={`/${locale}/register`} className="font-medium text-foreground hover:underline">
            {t('signUp')}
          </Link>
        </p>
      </CardContent>
    </Card>
  )
}

function LoginFormFallback() {
  return <div className="h-96" />
}

export default function LoginPage({ params }: { params: Promise<{ locale: Locale }> }) {
  return (
    <Suspense fallback={<LoginFormFallback />}>
      <LoginResolver params={params} />
    </Suspense>
  )
}

async function LoginResolver({ params }: { params: Promise<{ locale: Locale }> }) {
  const { locale } = await params
  return <LoginForm locale={locale} />
}
