<template>
  <svg
    class="sparkline"
    :viewBox="`0 0 ${width} ${height}`"
    :width="width"
    :height="height"
    preserveAspectRatio="none"
    role="img"
    aria-hidden="true"
  >
    <defs>
      <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" :stop-color="stroke" stop-opacity="0.28" />
        <stop offset="100%" :stop-color="stroke" stop-opacity="0" />
      </linearGradient>
    </defs>

    <path v-if="areaPath" :d="areaPath" :fill="`url(#${gradId})`" stroke="none" />
    <path
      v-if="linePath"
      :d="linePath"
      fill="none"
      :stroke="stroke"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
    <circle
      v-if="linePath"
      :cx="lastPoint.x"
      :cy="lastPoint.y"
      r="2.6"
      :fill="stroke"
    />
  </svg>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  data: { type: Array, default: () => [] },
  width: { type: Number, default: 120 },
  height: { type: Number, default: 36 },
  color: { type: String, default: '' },
  // 'auto' → 涨红跌绿（A股/加密惯例），或强制 'up' | 'down' | 'neutral'
  trend: { type: String, default: 'auto' },
})

let uid = 0
const gradId = `spark-${Date.now().toString(36)}-${(uid++).toString(36)}-${Math.floor(Math.random() * 1e6).toString(36)}`

const points = computed(() => {
  const arr = (props.data || []).map((v) => Number(v)).filter((v) => !Number.isNaN(v))
  if (arr.length < 2) return []
  const min = Math.min(...arr)
  const max = Math.max(...arr)
  const span = max - min || 1
  const pad = 3
  const innerW = props.width - pad * 2
  const innerH = props.height - pad * 2
  const step = innerW / (arr.length - 1)
  return arr.map((v, i) => ({
    x: pad + i * step,
    y: pad + innerH - ((v - min) / span) * innerH,
  }))
})

const direction = computed(() => {
  const arr = (props.data || []).map(Number).filter((v) => !Number.isNaN(v))
  if (arr.length < 2) return 'neutral'
  const delta = arr[arr.length - 1] - arr[0]
  if (delta > 0) return 'up'
  if (delta < 0) return 'down'
  return 'neutral'
})

const resolvedTrend = computed(() =>
  props.trend === 'auto' ? direction.value : props.trend
)

const stroke = computed(() => {
  if (props.color) return props.color
  if (resolvedTrend.value === 'up') return 'var(--up)'
  if (resolvedTrend.value === 'down') return 'var(--down)'
  return 'var(--text-secondary)'
})

const linePath = computed(() => {
  if (!points.value.length) return ''
  return points.value
    .map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(2)},${p.y.toFixed(2)}`)
    .join(' ')
})

const areaPath = computed(() => {
  if (!points.value.length) return ''
  const first = points.value[0]
  const last = points.value[points.value.length - 1]
  return `${linePath.value} L${last.x.toFixed(2)},${props.height} L${first.x.toFixed(2)},${props.height} Z`
})

const lastPoint = computed(() =>
  points.value.length ? points.value[points.value.length - 1] : { x: 0, y: 0 }
)
</script>

<style scoped>
.sparkline {
  display: block;
  overflow: visible;
}
</style>
