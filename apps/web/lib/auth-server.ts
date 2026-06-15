import { cookies } from 'next/headers'

export interface SessionUser {
  id: string
  email: string
  name: string
  role: 'customer' | 'admin'
}

const API_BASE = process.env.INTERNAL_API_URL ?? 'http://127.0.0.1:8080'

export async function getCurrentUser(): Promise<SessionUser | null> {
  const cookieStore = await cookies()
  const access = cookieStore.get('access_token')?.value
  if (!access) return null

  try {
    const res = await fetch(`${API_BASE}/api/v1/auth/me`, {
      headers: { Authorization: `Bearer ${access}` },
      cache: 'no-store',
    })
    if (!res.ok) return null
    return res.json()
  } catch {
    return null
  }
}
