import type { Config } from 'tailwindcss'

// All color tokens are stored in `globals.css` as space-separated OKLCH triplets
// (`L C H`), so Tailwind can wrap them in `oklch(... / <alpha>)` and still
// honor the `/50` opacity modifier that utility classes like `bg-primary/50`
// expect.
const ok = (name: string) => `oklch(var(--${name}))`

const config: Config = {
  darkMode: 'class',
  content: [
    './app/**/*.{ts,tsx}',
    './components/**/*.{ts,tsx}',
    './lib/**/*.{ts,tsx}',
  ],
  theme: {
    container: {
      center: true,
      padding: '1rem',
      screens: {
        '2xl': '1280px',
      },
    },
    extend: {
      colors: {
        border: ok('border'),
        input: ok('input'),
        ring: ok('ring'),
        background: ok('background'),
        foreground: ok('foreground'),
        primary: {
          DEFAULT: ok('primary'),
          foreground: ok('primary-foreground'),
          soft: ok('primary-soft'),
        },
        secondary: {
          DEFAULT: ok('secondary'),
          foreground: ok('secondary-foreground'),
        },
        destructive: {
          DEFAULT: ok('destructive'),
          foreground: ok('destructive-foreground'),
        },
        muted: {
          DEFAULT: ok('muted'),
          foreground: ok('muted-foreground'),
        },
        accent: {
          DEFAULT: ok('accent'),
          foreground: ok('accent-foreground'),
        },
        success: {
          DEFAULT: ok('success'),
          foreground: ok('success-foreground'),
          soft: ok('success-soft'),
        },
        warning: {
          DEFAULT: ok('warning'),
          foreground: ok('warning-foreground'),
          soft: ok('warning-soft'),
        },
        info: {
          DEFAULT: ok('info'),
          foreground: ok('info-foreground'),
          soft: ok('info-soft'),
        },
        card: {
          DEFAULT: ok('card'),
          foreground: ok('card-foreground'),
        },
        popover: {
          DEFAULT: ok('popover'),
          foreground: ok('popover-foreground'),
        },
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
      fontFamily: {
        sans: [
          'ui-sans-serif',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Inter',
          'Segoe UI',
          'Roboto',
          'sans-serif',
        ],
        mono: [
          'ui-monospace',
          'SFMono-Regular',
          '"SF Mono"',
          'Menlo',
          'Consolas',
          'monospace',
        ],
      },
      letterSpacing: {
        tightish: '-0.015em',
      },
    },
  },
  plugins: [],
}

export default config
