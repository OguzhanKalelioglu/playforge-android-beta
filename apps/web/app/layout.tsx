import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_SITE_URL ?? 'http://localhost:3000'),
  title: {
    default: 'TestersCommunity — Android Uygulama Test Hizmeti',
    template: '%s | TestersCommunity',
  },
  description:
    'Android uygulamanızı 25 gerçek kullanıcıyla 14 gün boyunca test edin. Detaylı rapor, canlı ilerleme.',
  keywords: ['android test', 'beta tester', 'google play', 'kapalı beta', 'uygulama testi'],
  authors: [{ name: 'TestersCommunity' }],
  creator: 'TestersCommunity',
  openGraph: {
    type: 'website',
    locale: 'tr_TR',
    siteName: 'TestersCommunity',
  },
  robots: {
    index: true,
    follow: true,
  },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="tr" suppressHydrationWarning>
      <body className="min-h-screen bg-background font-sans antialiased">
        {children}
      </body>
    </html>
  )
}
