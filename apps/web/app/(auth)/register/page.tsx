import Link from 'next/link'
import type { Metadata } from 'next'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export const metadata: Metadata = { title: 'Kayıt Ol' }

export default function RegisterPage() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Hesap Oluştur</CardTitle>
        <CardDescription>Ücretsiz hesabınızı oluşturun, hemen test satın alın</CardDescription>
      </CardHeader>
      <CardContent>
        <form className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Ad Soyad</Label>
            <Input id="name" name="name" type="text" placeholder="Ahmet Yılmaz" required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="email">E-posta</Label>
            <Input id="email" name="email" type="email" placeholder="ornek@firma.com" required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="password">Şifre</Label>
            <Input id="password" name="password" type="password" minLength={8} required />
            <p className="text-xs text-muted-foreground">En az 8 karakter, harf ve rakam içermeli</p>
          </div>
          <div className="flex items-start gap-2">
            <input id="terms" type="checkbox" required className="mt-1" />
            <Label htmlFor="terms" className="text-xs font-normal leading-relaxed text-muted-foreground">
              <Link href="/legal/terms" className="underline">Kullanım şartlarını</Link> ve{' '}
              <Link href="/legal/privacy" className="underline">gizlilik politikasını</Link> okudum, kabul ediyorum.
            </Label>
          </div>
          <Button type="submit" className="w-full">
            Hesap Oluştur
          </Button>
        </form>
        <div className="mt-6 text-center text-sm text-muted-foreground">
          Zaten hesabınız var mı?{' '}
          <Link href="/login" className="font-medium text-primary hover:underline">
            Giriş yapın
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
