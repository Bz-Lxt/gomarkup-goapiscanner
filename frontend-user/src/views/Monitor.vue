<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, wsURL, type Finding } from '../api'
import DefectTree from '../components/DefectTree.vue'
import Toast from '../components/Toast.vue'
import { useScanStore } from '../stores/scan'

interface WSEvent {
  type: string
  ts: string
  level?: string
  message?: string
  sent?: number
  total?: number
  hits?: number
  finding?: Finding
  status?: string
}

const route = useRoute()
const router = useRouter()
const store = useScanStore()
const logEl = ref<HTMLElement | null>(null)
const toast = ref('')
let ws: WebSocket | null = null
let poll: number | null = null

const id = computed(() => String(route.params.id || ''))
const percent = computed(() => {
  const t = store.task
  if (!t || !t.total) return 0
  return Math.min(100, Math.round((t.sent / t.total) * 100))
})

function applyEvent(ev: WSEvent) {
  if (ev.type === 'log') store.pushLine({ ts: ev.ts, level: ev.level || 'info', message: ev.message || '' })
  if (ev.type === 'progress' && store.task) {
    store.task.sent = ev.sent || store.task.sent
    store.task.total = ev.total || store.task.total
    store.task.hits = ev.hits || store.task.hits
  }
  if (ev.type === 'finding' && ev.finding) {
    store.pushFinding(ev.finding)
    store.pushLine({ ts: ev.ts, level: 'hit', message: `[${ev.finding.severity}] ${ev.finding.title} @ ${ev.finding.method} ${ev.finding.endpoint}` })
    if (store.task) store.task.hits = (store.task.hits || 0)
  }
  if (ev.type === 'done' && store.task) {
    store.task.status = ev.status || store.task.status
    store.pushLine({ ts: ev.ts, level: 'info', message: ev.message || `任务结束：${ev.status}` })
    void refreshTree()
  }
  nextTick(() => {
    if (logEl.value) logEl.value.scrollTop = logEl.value.scrollHeight
  })
}

async function refreshTree() {
  try {
    const pack = await api.findings(id.value)
    store.tree = pack.tree
    store.findings = pack.findings
    if (store.task) {
      store.task.critical = pack.stats.critical
      store.task.high = pack.stats.high
      store.task.medium = pack.stats.medium
      store.task.low = pack.stats.low
      store.task.info = pack.stats.info
      store.task.hits = pack.findings.length
    }
  } catch {
    /* keep stream */
  }
}

async function cancelScan() {
  try {
    await api.cancel(id.value)
  } catch (e) {
    toast.value = e instanceof Error ? e.message : '无法取消'
  }
}

onMounted(async () => {
  try {
    await store.hydrate(id.value)
  } catch (e) {
    toast.value = e instanceof Error ? e.message : '任务不存在'
    return
  }
  ws = new WebSocket(wsURL(id.value))
  ws.onmessage = (e) => {
    try {
      applyEvent(JSON.parse(e.data) as WSEvent)
    } catch { /* ignore */ }
  }
  poll = window.setInterval(async () => {
    try {
      store.task = await api.get(id.value)
      if (store.task.status === 'succeeded' || store.task.status === 'failed' || store.task.status === 'cancelled') {
        await refreshTree()
        if (poll) window.clearInterval(poll)
      }
    } catch { /* ignore */ }
  }, 1500)
})

onUnmounted(() => {
  ws?.close()
  if (poll) window.clearInterval(poll)
})
</script>

<template>
  <Toast :message="toast" @close="toast = ''" />
  <section v-if="store.task" class="w-full">
    <div class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="font-display text-3xl">漏洞监控看板</h1>
        <p class="mt-1 font-mono text-sm text-mute">{{ store.task.id }} · {{ store.task.base_url }}</p>
      </div>
      <div class="flex gap-3">
        <button class="rounded-lg border border-line px-4 py-2 text-sm" @click="cancelScan">停止扫描</button>
        <button class="rounded-lg bg-phosphor px-4 py-2 text-sm font-medium text-ink" @click="router.push(`/report/${id}`)">合规报告</button>
      </div>
    </div>

    <div class="mt-6 grid w-full gap-3 sm:grid-cols-2 xl:grid-cols-5">
      <div class="rounded-xl border border-line bg-panel p-4">
        <div class="text-xs text-mute">进度</div>
        <div class="font-display text-2xl">{{ percent }}%</div>
        <div class="mt-2 h-2 overflow-hidden rounded bg-ink">
          <div class="h-full bg-phosphor transition-all" :style="{ width: percent + '%' }" />
        </div>
      </div>
      <div class="rounded-xl border border-line bg-panel p-4">
        <div class="text-xs text-mute">已发 / 总量</div>
        <div class="font-display text-2xl">{{ store.task.sent }} / {{ store.task.total }}</div>
      </div>
      <div class="rounded-xl border border-crit/40 bg-panel p-4"><div class="text-xs text-mute">严重</div><div class="font-display text-2xl text-crit">{{ store.task.critical }}</div></div>
      <div class="rounded-xl border border-high/40 bg-panel p-4"><div class="text-xs text-mute">高危</div><div class="font-display text-2xl text-high">{{ store.task.high }}</div></div>
      <div class="rounded-xl border border-mid/40 bg-panel p-4"><div class="text-xs text-mute">中危 / 命中</div><div class="font-display text-2xl text-mid">{{ store.task.medium }} / {{ store.task.hits }}</div></div>
    </div>

    <div class="mt-6 grid w-full gap-6 xl:grid-cols-2">
      <div class="rounded-2xl border border-line bg-panel/80 p-5">
        <h2 class="mb-4 font-display text-lg">全生命周期缺陷树</h2>
        <DefectTree :nodes="store.tree" />
        <p v-if="!store.tree.length" class="text-sm text-mute">等待指纹命中…</p>
      </div>
      <div class="rounded-2xl border border-line bg-ink p-5">
        <h2 class="mb-4 font-display text-lg">漏洞实时打印流</h2>
        <div ref="logEl" class="h-[480px] overflow-auto font-mono text-xs leading-6">
          <div v-for="(ln, i) in store.lines" :key="i" class="border-b border-line/60 px-1" :class="ln.flash ? 'flash-row' : ''">
            <span class="text-mute">{{ ln.ts }}</span>
            <span class="mx-2" :class="ln.level === 'hit' ? 'text-crit' : 'text-phosphor'">{{ ln.level }}</span>
            <span>{{ ln.message }}</span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
