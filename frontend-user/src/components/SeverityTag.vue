<script setup lang="ts">
import type { Severity } from '../api'
const props = defineProps<{ severity?: Severity; label?: string }>()
const map: Record<string, { text: string; cls: string }> = {
  critical: { text: '严重', cls: 'border-crit text-crit bg-crit/10' },
  high: { text: '高危', cls: 'border-high text-high bg-high/10' },
  medium: { text: '中危', cls: 'border-mid text-mid bg-mid/10' },
  low: { text: '低危', cls: 'border-low text-low bg-low/10' },
  info: { text: '信息', cls: 'border-mute text-mute bg-white/5' },
}
const item = map[props.severity || 'info']
</script>

<template>
  <span class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 font-display text-[11px] tracking-wide" :class="item.cls">
    <i class="h-1.5 w-1.5 rounded-full bg-current" />
    {{ label || item.text }}
  </span>
</template>
