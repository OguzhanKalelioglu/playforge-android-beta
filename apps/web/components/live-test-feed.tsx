import { Activity, CheckCircle2, Clock, Smartphone } from 'lucide-react'

// Realistic data points. Intentionally not "lorem ipsum" — the same shape
// the orchestrator will return. Static for now; the data shape is what
// matters here, not the source.
const FEED_ROWS = [
  { emulator: 'EM-001 · Pixel 6', action: 'opt_in',     status: 'success', ago: '0:24 ago' },
  { emulator: 'EM-007 · Galaxy A52', action: 'install',  status: 'success', ago: '0:38 ago' },
  { emulator: 'EM-003 · Pixel 4a',   action: 'open',      status: 'success', ago: '1:02 ago' },
  { emulator: 'EM-012 · Redmi Note', action: 'interact',  status: 'success', ago: '1:14 ago' },
  { emulator: 'EM-002 · Pixel 7',    action: 'opt_in',     status: 'success', ago: '1:27 ago' },
  { emulator: 'EM-009 · Galaxy S21', action: 'install',   status: 'success', ago: '1:41 ago' },
] as const

const ACTION_LABEL: Record<(typeof FEED_ROWS)[number]['action'], string> = {
  opt_in: 'Opt-in completed',
  install: 'App installed',
  open: 'App opened',
  interact: 'Session 12m 04s',
}

export function LiveTestFeed() {
  return (
    <div className="relative overflow-hidden rounded-2xl border bg-card text-card-foreground shadow-sm">
      {/* status header — the live indicator is the proof object */}
      <div className="flex items-center justify-between border-b bg-muted/40 px-5 py-3">
        <div className="flex items-center gap-2.5">
          <span className="relative flex h-2.5 w-2.5">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-success/60" />
            <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-success" />
          </span>
          <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Live test · day 3 of 14
          </span>
        </div>
        <span className="mono text-xs text-muted-foreground tabular-nums">
          18 / 25 active
        </span>
      </div>

      <ul className="divide-y">
        {FEED_ROWS.map((row, i) => (
          <li
            key={`${row.emulator}-${i}`}
            className="grid grid-cols-[1fr_auto] items-center gap-3 px-5 py-3 text-sm"
          >
            <div className="flex min-w-0 items-center gap-3">
              <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
                <Smartphone className="h-3.5 w-3.5" strokeWidth={1.75} />
              </span>
              <div className="min-w-0">
                <div className="truncate font-mono text-xs text-muted-foreground">
                  {row.emulator}
                </div>
                <div className="mt-0.5 flex items-center gap-1.5 text-sm">
                  {row.status === 'success' ? (
                    <CheckCircle2
                      className="h-3.5 w-3.5 shrink-0 text-success"
                      strokeWidth={2}
                    />
                  ) : (
                    <Activity className="h-3.5 w-3.5 shrink-0 text-warning" strokeWidth={2} />
                  )}
                  <span className="truncate">{ACTION_LABEL[row.action]}</span>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground tabular-nums">
              <Clock className="h-3 w-3" strokeWidth={1.75} />
              {row.ago}
            </div>
          </li>
        ))}
      </ul>

      <div className="flex items-center justify-between border-t bg-muted/30 px-5 py-2.5 text-xs text-muted-foreground">
        <span>com.example.app · closed beta</span>
        <span className="font-mono tabular-nums">~ 24 events / hour</span>
      </div>
    </div>
  )
}
