<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Task } from '../api'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import Toast from '../components/Toast.vue'

const router = useRouter()
const metaURL = ref('http://localhost:28483')
const form = reactive({
  base_url: '',
  concurrency: 16,
  timeout_ms: 5000,
  authorized: false,
})
const errors = reactive({ base_url: '', authorized: '', swagger: '' })
const file = ref<File | null>(null)
const dragging = ref(false)
const confirmOpen = ref(false)
const toast = ref('')
const recent = ref<Task[]>([])

function urlOK(s: string) {
  try {
    const u = new URL(s)
    return u.protocol === 'http:' || u.protocol === 'https:'
  } catch {
    return false
  }
}

function validate(): boolean {
  errors.base_url = ''
  errors.authorized = ''
  errors.swagger = ''
  const base = form.base_url.trim() || metaURL.value
  if (!urlOK(base)) errors.base_url = '请输入合法的 http(s) URL，例如靶场地址'
  if (form.concurrency < 1 || form.concurrency > 64) errors.base_url = errors.base_url || '并发需在 1–64'
  if (form.timeout_ms < 1000) errors.base_url = errors.base_url || '超时至少 1000ms'
  if (!form.authorized) errors.authorized = '必须勾选授权声明'
  if (file.value && !file.value.name.toLowerCase().endsWith('.json')) {
    errors.swagger = '仅接受 Swagger / OpenAPI JSON'
  }
  return !errors.base_url && !errors.authorized && !errors.swagger
}

function onFile(f?: File) {
  file.value = f || null
}

async function submit() {
  if (!validate()) {
    toast.value = '请先修正表单错误'
    return
  }
  confirmOpen.value = true
}

async function doCreate() {
  confirmOpen.value = false
  try {
    const base = form.base_url.trim() || metaURL.value
    let task: Task
    if (file.value) {
      const fd = new FormData()
      fd.set('base_url', base)
      fd.set('concurrency', String(form.concurrency))
      fd.set('timeout_ms', String(form.timeout_ms))
      fd.set('authorized', 'true')
      fd.set('swagger', file.value)
      task = await api.createForm(fd)
    } else {
      task = await api.createJSON({
        base_url: base,
        concurrency: form.concurrency,
        timeout_ms: form.timeout_ms,
        authorized: true,
      })
    }
    router.push(`/monitor/${task.id}`)
  } catch (e) {
    toast.value = e instanceof Error ? e.message : '启动失败'
  }
}

onMounted(async () => {
  try {
    const m = await api.meta()
    metaURL.value = m.default_base_url
    if (!form.base_url) form.base_url = m.default_base_url
    recent.value = await api.list()
  } catch {
    form.base_url = metaURL.value
  }
})
</script>

<template>
  <Toast :message="toast" @close="toast = ''" />
  <ConfirmDialog
    :open="confirmOpen"
    title="确认对授权目标启动扫描"
    body="系统将向目标发送带有检测载荷的变异请求。请确认你拥有该目标或已获书面授权。默认演示请使用内置靶场。"
    @cancel="confirmOpen = false"
    @confirm="doCreate"
  />

  <section class="w-full">
    <h1 class="font-display text-3xl font-semibold">安全扫描控制台</h1>
    <p class="mt-2 text-mute">输入基础 URL，或上传 Swagger JSON。扫描只应针对已授权资产；演示请使用内置靶场。</p>

    <div class="mt-8 grid w-full gap-6 lg:grid-cols-5">
      <form class="lg:col-span-3 space-y-5 rounded-2xl border border-line bg-panel/90 p-6" @submit.prevent="submit">
        <div>
          <label class="text-sm text-mute">基础 URL <span class="text-crit">*</span></label>
          <input v-model="form.base_url" class="mt-2 w-full rounded-xl border border-line bg-ink px-4 py-3 outline-none focus:border-phosphor" placeholder="http://localhost:28483" />
          <p v-if="errors.base_url" class="mt-1 text-sm text-crit">{{ errors.base_url }}</p>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="text-sm text-mute">并发数</label>
            <input v-model.number="form.concurrency" type="number" min="1" max="64" class="mt-2 w-full rounded-xl border border-line bg-ink px-4 py-3" />
          </div>
          <div>
            <label class="text-sm text-mute">超时 (ms)</label>
            <input v-model.number="form.timeout_ms" type="number" min="1000" class="mt-2 w-full rounded-xl border border-line bg-ink px-4 py-3" />
          </div>
        </div>
        <div
          class="rounded-xl border border-dashed border-line px-4 py-8 text-center"
          :class="dragging ? 'border-phosphor bg-phosphor/5' : ''"
          @dragover.prevent="dragging = true"
          @dragleave="dragging = false"
          @drop.prevent="dragging = false; onFile($event.dataTransfer?.files?.[0])"
        >
          <p class="text-sm text-mute">拖拽上传 Swagger / OpenAPI JSON（可选）</p>
          <input class="mt-3 text-sm" type="file" accept="application/json,.json" @change="onFile(($event.target as HTMLInputElement).files?.[0] || undefined)" />
          <p v-if="file" class="mt-2 font-mono text-xs text-phosphor">{{ file.name }}</p>
          <p v-if="errors.swagger" class="mt-1 text-sm text-crit">{{ errors.swagger }}</p>
        </div>
        <label class="flex items-start gap-3 text-sm">
          <input v-model="form.authorized" type="checkbox" class="mt-1" />
          <span>我确认目标为自有或已授权资产，并了解扫描将发送检测载荷（内置靶场已预授权演示）。</span>
        </label>
        <p v-if="errors.authorized" class="text-sm text-crit">{{ errors.authorized }}</p>
        <button class="w-full rounded-xl bg-phosphor py-3 font-display font-semibold text-ink hover:brightness-110">启动扫描</button>
      </form>

      <aside class="lg:col-span-2 rounded-2xl border border-line bg-panel/80 p-6">
        <h2 class="font-display text-lg">最近任务</h2>
        <ul class="mt-4 space-y-3">
          <li v-for="t in recent" :key="t.id" class="rounded-xl border border-line px-3 py-3">
            <div class="flex items-center justify-between">
              <span class="font-mono text-xs text-mute">{{ t.created_at }}</span>
              <span class="text-xs text-phosphor">{{ t.status }}</span>
            </div>
            <button class="mt-1 text-left text-sm hover:text-phosphor" @click="router.push(`/monitor/${t.id}`)">{{ t.base_url }}</button>
          </li>
          <li v-if="!recent.length" class="text-sm text-mute">暂无任务</li>
        </ul>
      </aside>
    </div>
  </section>
</template>
