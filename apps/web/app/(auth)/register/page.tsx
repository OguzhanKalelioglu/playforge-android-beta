'use client'

import { useState, Suspense } from 'react'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const schema = z
  .object({
    name: z.string().min(2, 'Ad en az 2 karakter olmalı.'),
    email: z.string().email('Geçerli bir e-posta girin.'),
    password: z
      .string()
      .min(8, 'Şifre en az 8 karakter olmalı.')
      .regex(/[A-Za-z]/, 'En az bir harf içermeli.')
      .regex(/[0-9]/, 'En az bir rakam içermeli.'),
    acceptTerms: z.literal(true, {
      errorMap: () => ({ message: 'Şartları kabul etmelisin.' }),
    }),
  })
type FormValues = z.infer<typeof schema>

function RegisterForm() {
  const router = useRouter()
  const search = useSearchParams()
  const plan = search.get('plan')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

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
          setError('Bu e-posta zaten kayıtlı. Şifreni mi unuttun?')
          return
        }
        setError(body.error ?? 'Kayıt başarısız.')
        return
      }
      router.push(plan ? `/dashboard/new?plan=${plan}` : '/dashboard')
    } catch {
      setError('Bağlantı hatası. Tekrar deneyin.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Hesap oluştur</CardTitle>
        <CardDescription>Ücretsiz. Ödeme sonra.</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">Ad Soyad</Label>
            <Input
              id="name"
              autoComplete="name"
              placeholder="Ahmet Yılmaz"
              aria-invalid={!!errors.name}
              {...register('name')}
            />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="email">E-posta</Label>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              placeholder="ornek@firma.com"
              aria-invalid={!!errors.email}
              {...register('email')}
            />
            {errors.email && <p className="text-xs text-destructive">{errors.email.message}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">Şifre</Label>
            <Input
              id="password"
              type="password"
              autoComplete="new-password"
              aria-invalid={!!errors.password}
              {...register('password')}
            />
            <p className="text-xs text-muted-foreground">En az 8 karakter, harf ve rakam içermeli.</p>
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
              <Link href="/legal/terms" className="underline">
                Kullanım şartlarını
              </Link>{' '}
              ve{' '}
              <Link href="/legal/privacy" className="underline">
                gizlilik politikasını
              </Link>{' '}
              okudum, kabul ediyorum.
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
            {submitting ? 'Oluşturuluyor…' : 'Hesap oluştur'}
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          Zaten hesabın var mı?{' '}
          <Link href="/login" className="font-medium text-foreground hover:underline">
            Giriş yap
          </Link>
        </p>
      </CardContent>
    </Card>
  )
}

export default function RegisterPage() {
  return (
    <Suspense fallback={null}>
      <RegisterForm />
    </Suspense>
  )
}
