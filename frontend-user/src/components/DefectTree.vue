<script setup lang="ts">
import { ref } from 'vue'
import type { DefectNode } from '../api'
import SeverityTag from './SeverityTag.vue'

defineProps<{ nodes: DefectNode[] }>()
const open = ref<Record<string, boolean>>({})
function toggle(key: string) {
  open.value[key] = !open.value[key]
}
</script>

<template>
  <ul class="space-y-2">
    <li v-for="n in nodes" :key="n.key" class="rounded-xl border border-line bg-ink/40">
      <button class="flex w-full items-center justify-between px-4 py-3 text-left" @click="toggle(n.key)">
        <div class="flex items-center gap-3">
          <span class="font-mono text-xs text-phosphor">{{ n.method }}</span>
          <span class="font-medium">{{ n.path || n.label }}</span>
          <SeverityTag v-if="n.severity" :severity="n.severity" />
        </div>
        <span class="text-xs text-mute">{{ n.count }} 项 · {{ open[n.key] ? '收起' : '展开' }}</span>
      </button>
      <div v-if="open[n.key]" class="space-y-2 border-t border-line px-4 py-3">
        <div v-for="c in n.children" :key="c.key" class="rounded-lg bg-panel/80 p-3">
          <div class="flex flex-wrap items-center gap-2">
            <SeverityTag :severity="c.severity" />
            <span class="font-display">{{ c.label }}</span>
          </div>
          <p v-if="c.finding" class="mt-2 font-mono text-xs text-mute">{{ c.finding.evidence }}</p>
          <p v-if="c.finding" class="mt-1 font-mono text-xs text-phosphor/80">payload: {{ c.finding.payload }}</p>
        </div>
      </div>
    </li>
  </ul>
</template>
