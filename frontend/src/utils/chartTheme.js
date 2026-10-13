// ECharts theme helpers — keeps charts in sync with the app's light/dark theme.
import { isDark } from './theme'

// Read a CSS custom property from :root so charts inherit the design tokens.
export function token(name, fallback = '') {
  if (typeof window === 'undefined') return fallback
  const v = getComputedStyle(document.documentElement).getPropertyValue(name)
  return (v && v.trim()) || fallback
}

// Build a theme-aware palette + axis/tooltip config for candlestick charts.
export function getChartTheme() {
  const dark = isDark.value
  return {
    dark,
    textColor: token('--text-muted', dark ? '#94a3b8' : '#64748b'),
    textMain: token('--text-main', dark ? '#f1f5f9' : '#0f172a'),
    lineColor: token('--border-color', dark ? 'rgba(148,163,184,0.14)' : 'rgba(100,116,139,0.2)'),
    splitColor: token('--border-subtle', dark ? 'rgba(148,163,184,0.09)' : 'rgba(100,116,139,0.12)'),
    upColor: token('--up', '#ef4444'),
    downColor: token('--down', '#10b981'),
    primary: token('--primary', '#3b82f6'),
    tooltipBg: dark ? 'rgba(17, 24, 39, 0.95)' : 'rgba(255, 255, 255, 0.98)',
    tooltipBorder: dark ? 'rgba(148,163,184,0.22)' : 'rgba(100,116,139,0.18)',
    tooltipText: dark ? '#e2e8f0' : '#0f172a',
    crosshair: dark ? 'rgba(148,163,184,0.5)' : 'rgba(100,116,139,0.45)',
    crosshairLabelBg: dark ? '#1e293b' : '#334155',
    volumeUp: 'rgba(239, 68, 68, 0.35)',
    volumeDown: 'rgba(16, 185, 129, 0.35)',
  }
}

// Shared tooltip style object for ECharts options.
export function tooltipStyle(t) {
  return {
    backgroundColor: t.tooltipBg,
    borderColor: t.tooltipBorder,
    borderWidth: 1,
    padding: [10, 12],
    textStyle: { color: t.tooltipText, fontSize: 12 },
    extraCssText: 'border-radius:10px; box-shadow:0 12px 32px -8px rgba(15,23,42,0.35);',
  }
}

// Crosshair pointer styling.
export function crosshairStyle(t) {
  return {
    lineStyle: { color: t.crosshair, width: 1, type: 'dashed' },
    crossStyle: { color: t.crosshair },
    label: {
      backgroundColor: t.crosshairLabelBg,
      borderRadius: 6,
      color: '#fff',
      padding: [4, 8],
      fontSize: 11,
    },
  }
}
