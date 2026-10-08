// Theme management utility — light / dark toggle with persistence
import { ref } from 'vue'

const THEME_KEY = 'cex-theme'
export const isDark = ref(false)

// Apply a given darkness state to the document root
export function applyTheme(dark) {
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
  document.documentElement.classList.toggle('light', !dark)
}

// Read saved preference / OS default and apply
export function initTheme() {
  let dark = false
  const saved = localStorage.getItem(THEME_KEY)
  if (saved === 'dark' || saved === 'light') {
    dark = saved === 'dark'
  } else if (typeof window !== 'undefined') {
    dark = window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  applyTheme(dark)
}

// Toggle and persist
export function toggleTheme() {
  const next = !isDark.value
  applyTheme(next)
  localStorage.setItem(THEME_KEY, next ? 'dark' : 'light')
}
