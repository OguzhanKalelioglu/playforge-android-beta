const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? 'http://127.0.0.1:8080'

export interface PlanTier {
  id: string
  slug: string
  name: string
  description: string
  tester_count: number
  duration_days: number
  price_try: number
  price_usd: number | null
  features: string[]
  sort_order: number
}

export interface Test {
  id: string
  package_name: string
  test_link: string | null
  status: 'pending' | 'active' | 'completed' | 'failed' | 'cancelled'
  started_at: string | null
  ends_at: string | null
  created_at: string
  progress?: {
    total: number
    opt_in: number
    installed: number
    engaged: number
    reviewed: number
    failed: number
  }
}

export interface Assignment {
  id: string
  tester_email: string
  status: 'pending' | 'in_progress' | 'completed' | 'failed' | 'skipped'
  opt_in_at: string | null
  install_at: string | null
  last_engagement_at: string | null
  error_message: string | null
}

export interface ActivityEvent {
  id: number
  action: string
  performed_at: string
  success: boolean
  error_message: string | null
  screenshot_path: string | null
  metadata: Record<string, unknown>
}

export interface Order {
  id: string
  plan_slug: string
  plan_name: string
  status: 'pending' | 'awaiting_payment' | 'paid' | 'failed' | 'cancelled' | 'refunded'
  total: number
  currency: string
  created_at: string
  paid_at: string | null
  expires_at: string
}

export interface Review {
  id: string
  rating: number
  review_text: string
  language: string
  posted_at: string
}

export class ApiError extends Error {
  constructor(public status: number, public body: unknown) {
    super(`API error ${status}`)
  }
}

async function request<T>(path: string, init?: RequestInit, token?: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
    cache: 'no-store',
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body)
  }
  return res.json()
}

export const api = {
  plans: () => request<PlanTier[]>('/api/v1/plans'),

  tests: (token: string) => request<Test[]>('/api/v1/tests', undefined, token),
  test: (id: string, token: string) =>
    request<Test & { assignments: Assignment[] }>(`/api/v1/tests/${id}`, undefined, token),
  testActivity: (id: string, token: string) =>
    request<ActivityEvent[]>(`/api/v1/tests/${id}/activity`, undefined, token),
  testReviews: (id: string, token: string) =>
    request<Review[]>(`/api/v1/tests/${id}/reviews`, undefined, token),

  orders: (token: string) => request<Order[]>('/api/v1/orders', undefined, token),
  createOrder: (
    data: { plan_slug: string; package_name: string; test_link: string },
    token: string
  ) =>
    request<Order>('/api/v1/orders', { method: 'POST', body: JSON.stringify(data) }, token),
  order: (id: string, token: string) =>
    request<Order & { payment_url?: string }>(`/api/v1/orders/${id}`, undefined, token),
}
