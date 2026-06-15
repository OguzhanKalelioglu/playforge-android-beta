import { getCurrentUser } from '@/lib/auth-server'
import { redirect } from 'next/navigation'
import { User, Mail, Calendar, type LucideIcon } from 'lucide-react'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { formatDate } from '@/lib/format'

export const metadata = { title: 'Hesap Ayarları' }

export default async function SettingsPage() {
  const user = await getCurrentUser()
  if (!user) redirect('/login?next=/dashboard/settings')

  return (
    <div className="container max-w-3xl py-8">
      <h1 className="text-2xl font-semibold tracking-tightish">Hesap ayarları</h1>
      <p className="mt-1 text-sm text-muted-foreground">Profil bilgilerin ve hesap detayların.</p>

      <div className="mt-8 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Profil</CardTitle>
            <CardDescription>Ad ve e-posta bilgilerin.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 text-sm">
            <Row icon={User} label="Ad Soyad" value={user.name} />
            <Row icon={Mail} label="E-posta" value={user.email} />
            <Row
              icon={Calendar}
              label="Hesap rolü"
              value={
                <Badge variant={user.role === 'admin' ? 'info' : 'muted'}>
                  {user.role === 'admin' ? 'Yönetici' : 'Müşteri'}
                </Badge>
              }
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Güvenlik</CardTitle>
            <CardDescription>Şifreni değiştir ve aktif oturumları yönet.</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Şifre değiştirme ve oturum yönetimi yakında. Şimdilik destek için iletişime geç.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function Row({
  icon: Icon,
  label,
  value,
}: {
  icon: LucideIcon
  label: string
  value: React.ReactNode
}) {
  return (
    <div className="flex items-center justify-between border-b pb-3 last:border-0 last:pb-0">
      <div className="flex items-center gap-2 text-muted-foreground">
        <Icon className="h-4 w-4" />
        <span>{label}</span>
      </div>
      <span className="font-medium">{value}</span>
    </div>
  )
}
