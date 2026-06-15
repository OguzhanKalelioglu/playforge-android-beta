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

function RegisterForm({ locale }: { locale: Locale }) {
  const t = useTranslations('auth')
  const tLegal = useTranslations('footer')
  const router = useRouter()
  const search = useSearchParams()
  const plan = search.get('plan')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const schema = z
    .object({
      name: z.string().min(2, t('nameRequired')),
      email: z.string().email(t('emailInvalid')),
      password: z
        .string()
        .min(8, t('passwordTooShort'))
        .regex(/[A-Za-z]/, t('passwordHint'))
        .regex(/[0-9]/, t('passwordHint')),
      acceptTerms: z.literal(true, {
        errorMap: () => ({ message: t('nameRequired') }),
      }),
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
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        if (body.error?.includes('email')) {
          setError(t('loginError'))
          return
        }
        setError(body.error ?? t('registerError'))
        return
      }
      router.push(plan ? `/${locale}/dashboard/new?plan=${plan}` : `/${locale}/dashboard`)
    } catch {
      setError(t('registerError'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('registerTitle')}</CardTitle>
        <CardDescription>{t('registerSubtitle')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">{t('name')}</Label>
            <Input
              id="name"
              autoComplete="name"
              placeholder={t('namePlaceholder')}
              aria-invalid={!!errors.name}
              {...register('name')}
            />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>

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
            <Label htmlFor="password">{t('password')}</Label>
            <Input
              id="password"
              type="password"
              autoComplete="new-password"
              aria-invalid={!!errors.password}
              placeholder={t('passwordPlaceholder')}
              {...register('password')}
            />
            <p className="text-xs text-muted-foreground">{t('passwordHint')}</p>
            {errors.password && (
              <p className="text-xs text-destructive">{errors.password.message}</p>
            )}
          </div>

          <label className="flex items-start gap-2 text-sm">
            <input
              type="checkbox"
              className="mt-0.5 h-4 w-4 rounded border-input"
              aria-invalid={!!errors.acceptTerms}
              {...register('acceptTerms')}
            />
            <span className="text-muted-foreground">
              <Link href={`/${locale}/legal/terms`} className="underline">
                {tLegal('legalTerms')}
              </Link>{' '}
              ·{' '}
              <Link href={`/${locale}/legal/privacy`} className="underline">
                {tLegal('legalPrivacy')}
              </Link>
            </span>
          </label>
          {errors.acceptTerms && (
            <p className="text-xs text-destructive">{errors.acceptTerms.message}</p>
          )}

          {error && (
            <div role="alert" className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          )}

          <Button type="submit" className="w-full" disabled={submitting}>
            {submitting ? t('creating') : t('register')}
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          {t('hasAccount')}{' '}
          <Link href={`/${locale}/login`} className="font-medium text-foreground hover:underline">
            {t('signIn')}
          </Link>
        </p>
      </CardContent>
    </Card>
  )
}

function RegisterFormFallback() {
  return <div className="h-96" />
}

export default function RegisterPage({ params }: { params: Promise<{ locale: Locale }> }) {
  return (
    <Suspense fallback={<RegisterFormFallback />}>
      <RegisterResolver params={params} />
    </Suspense>
  )
}

async function RegisterResolver({ params }: { params: Promise<{ locale: Locale }> }) {
  const { locale } = await params
  return <RegisterForm locale={locale} />
}
