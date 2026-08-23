<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, pdfURL, type ReportPreview } from '../api'
import SeverityTag from '../components/SeverityTag.vue'
import Toast from '../components/Toast.vue'

const route = useRoute()
const router = useRouter()
const data = ref<ReportPreview | null>(null)
const toast = ref('')

onMounted(async () => {
  try {
    data.value = await api.report(String(route.params.id))
  } catch (e) {
    toast.value = e instanceof Error ? e.message : '报告加载失败'
  }
})
</script>

<template>
  <Toast :message="toast" @close="toast = ''" />
  <article v-if="data" class="w-full">
    <div class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <p class="font-display text-phosphor">COMPLIANCE DOSSIER</p>
        <h1 class="mt-1 font-display text-3xl">安全合规报告预览</h1>
        <p class="mt-1 text-sm text-mute">生成时间 {{ data.generated_at }} · 任务 {{ data.task.id }}</p>
      </div>
      <div class="flex gap-3">
        <button class="rounded-lg border border-line px-4 py-2 text-sm" @click="router.push(`/monitor/${data.task.id}`)">返回看板</button>
        <a class="rounded-lg bg-phosphor px-4 py-2 text-sm font-medium text-ink" :href="pdfURL(data.task.id)">下载 PDF</a>
      </div>
    </div>

    <section class="mt-8 rounded-2xl border border-line bg-panel p-8">
      <h2 class="font-display text-xl">封面摘要</h2>
      <p class="mt-3 text-mute">目标 {{ data.task.base_url }}，状态 {{ data.task.status }}，共发送 {{ data.task.sent }} 个变异请求，命中 {{ data.task.hits }} 处指纹。</p>
      <div class="mt-6 grid gap-3 sm:grid-cols-5">
        <div class="rounded-xl bg-ink p-4"><SeverityTag severity="critical" /><div class="mt-2 font-display text-2xl">{{ data.stats.critical }}</div></div>
        <div class="rounded-xl bg-ink p-4"><SeverityTag severity="high" /><div class="mt-2 font-display text-2xl">{{ data.stats.high }}</div></div>
        <div class="rounded-xl bg-ink p-4"><SeverityTag severity="medium" /><div class="mt-2 font-display text-2xl">{{ data.stats.medium }}</div></div>
        <div class="rounded-xl bg-ink p-4"><SeverityTag severity="low" /><div class="mt-2 font-display text-2xl">{{ data.stats.low }}</div></div>
        <div class="rounded-xl bg-ink p-4"><SeverityTag severity="info" /><div class="mt-2 font-display text-2xl">{{ data.stats.info }}</div></div>
      </div>
    </section>

    <section class="mt-6 rounded-2xl border border-line bg-panel p-8">
      <h2 class="font-display text-xl">代码修复建议</h2>
      <ol class="mt-4 list-decimal space-y-3 pl-5 text-sm leading-7">
        <li v-for="(a, i) in data.advice" :key="i">{{ a }}</li>
      </ol>
    </section>

    <section class="mt-6 rounded-2xl border border-line bg-panel p-8">
      <h2 class="font-display text-xl">漏洞明细</h2>
      <div class="mt-4 overflow-x-auto">
        <table class="w-full min-w-[720px] text-left text-sm">
          <thead class="text-mute">
            <tr>
              <th class="py-2">级别</th>
              <th>接口</th>
              <th>标题</th>
              <th>建议</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="f in data.findings" :key="f.id" class="border-t border-line align-top">
              <td class="py-3"><SeverityTag :severity="f.severity" /></td>
              <td class="py-3 font-mono text-xs">{{ f.method }} {{ f.endpoint }}</td>
              <td class="py-3">{{ f.title }}<div class="mt-1 font-mono text-xs text-mute">{{ f.evidence }}</div></td>
              <td class="py-3 text-mute">{{ f.advice }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </article>
</template>
